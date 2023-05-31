package app

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/net/html"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/database"
	"github.com/anthony-ozdemir/zfse/internal/enum"
	"github.com/anthony-ozdemir/zfse/internal/filebuf"
	"github.com/anthony-ozdemir/zfse/internal/helper"
	"github.com/anthony-ozdemir/zfse/internal/path_manager"
)

func (a *Application) getTotalPreCrawlFileLinesToRead() (map[string]int, error) {
	// TODO [HP]: Performance benchmark with a +30GB file. If performance is too bad then
	// we need to estimate via read bytes. Though, this would increase code complexity.
	totalLinesMap := make(map[string]int)
	for zoneName, _ := range a.zoneFileRegistry {
		preCrawlCacheFile := path_manager.GetPreCrawlFilterOutputFilePath(zoneName)
		lines, err := helper.CountLinesInFile(preCrawlCacheFile)
		if err != nil {
			return nil, err
		}
		totalLinesMap[zoneName] = lines
	}
	return totalLinesMap, nil
}

func (a *Application) runPostCrawlFilters(wg *sync.WaitGroup) {
	defer wg.Done()

	// Let's first setup work item estimates
	totalLinesMap, err := a.getTotalPreCrawlFileLinesToRead()
	if err != nil {
		zap.L().Fatal("Unable to read zone file lines.")
	}
	totalWorkItems := 0
	for _, lines := range totalLinesMap {
		totalWorkItems += lines
	}
	a.applicationStateManager.SetTotalWorkItems(totalWorkItems)
	var processedWorkItemsMutex sync.Mutex
	processedWorkItems := 0
	a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)

	// Design Note:
	// 1. Read the input pre-crawl cache file line by line.
	// 2. Crawl the domain index pages one by one.
	// 2. Record the header and body during crawl.
	// 3. Feed domain properties through the post-crawl filters one by one.
	// 4. Append the output DomainProperties struct to post-crawl filter cache file.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Design Note: We will use a buffered channel to control the concurrent
	// crawler usage
	availableConnectors := make(chan struct{}, a.config.GeneralOptions.ConcurrentConnections)
	for i := 0; i < a.config.GeneralOptions.ConcurrentConnections; i++ {
		availableConnectors <- struct{}{}
	}

	var appendedToFileMutex sync.RWMutex
	bHasAppendedToFileOnce := false
	for zoneName, _ := range a.zoneFileRegistry {
		// Check the state of pre-crawl
		taskState := a.db.GetPostCrawlFilterTaskState(zoneName)
		if taskState.BIsFinished {
			appendedToFileMutex.Lock()
			bHasAppendedToFileOnce = true
			appendedToFileMutex.Unlock()

			processedWorkItemsMutex.Lock()
			defer processedWorkItemsMutex.Unlock()
			processedWorkItems += totalLinesMap[zoneName]
			a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
			continue
		}

		// Open the cache file from pre-connection filter
		preCrawlCacheFile := path_manager.GetPreCrawlFilterOutputFilePath(zoneName)

		file, err := os.Open(preCrawlCacheFile)
		if err != nil {
			zap.L().Fatal("Error opening file.", zap.String("err", err.Error()))
		}
		defer file.Close()

		startLineIndex := taskState.LineIndex
		lineIndex := 0

		// Prepare output file buffer
		outputFileBufferOpts := filebuf.FileOutputBufferOptions{
			BulkOutputLimit: a.config.GeneralOptions.FileBulkOutputQty,
			FilePath:        path_manager.GetPostCrawlFilterOutputFilePath(zoneName),
			OnAppendCB: func() {
				processedWorkItemsMutex.Lock()
				defer processedWorkItemsMutex.Unlock()
				processedWorkItems++
				a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
			},
			OnFlushCB: func(appendQty int) {
				// Design Note: This function will be called from a different thread.
				// Thus, we need to rely on appendQty of fileOutputBuffer.
				// TODO [LP]: Is this correct? Check via an unit test.
				lastIndex := startLineIndex + appendQty
				// Save Task State at this point
				taskState := database.PreCrawlFilterTaskState{
					BIsFinished: false,
					LineIndex:   lastIndex,
				}
				a.db.SavePreCrawlFilterTaskState(zoneName, taskState)
			},
		}
		fileOutputBuffer := filebuf.NewFileOutputBuffer(outputFileBufferOpts)

		// Create a bufio.Scanner to read the file line by line
		reader := bufio.NewReader(file)

		for {
			// Check if we need to gracefully shut-down before finishing this task.
			currentState := a.applicationStateManager.GetApplicationState()
			if currentState.task == enum.Shutdown {
				cancel()

				// Let's wait until all connectors are complete
				for i := 0; i < a.config.GeneralOptions.ConcurrentConnections; i++ {
					<-availableConnectors
				}

				fileOutputBuffer.Flush()

				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				zap.L().Fatal(
					"Error reading a line from pre-crawl cache file.",
					zap.String("err", err.Error()),
				)
			}

			if lineIndex >= startLineIndex {
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					// Skip empty lines
					processedWorkItemsMutex.Lock()
					defer processedWorkItemsMutex.Unlock()
					processedWorkItems++
					a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
					continue
				}

				// All lines are supposed to be JSON objects at this stage
				domainProperties := common.DomainProperties{}
				err := json.Unmarshal([]byte(line), &domainProperties)
				if err != nil {
					zap.L().Fatal("Unable to parse JSON.", zap.String("err", err.Error()))
				}

				// Check if we can launch a new connector
				<-availableConnectors // Acquire a connector

				// Launch a new connector
				// Design Note: Ensure to copy domainProperties otherwise it will cause
				// data race issues.
				go func(ctx context.Context, domainProperties common.DomainProperties) {
					defer func() { availableConnectors <- struct{}{} }() // Release a connector

					bHasDNSRecord := a.crawler.HasDNSARecord(ctx, domainProperties.DomainName)
					if !bHasDNSRecord {
						processedWorkItemsMutex.Lock()
						defer processedWorkItemsMutex.Unlock()
						processedWorkItems++
						a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
						return
					}

					url := a.config.GeneralOptions.ConnectionProtocol + "://" + domainProperties.DomainName

					bCanCrawl := a.crawler.CanCrawl(ctx, url)
					if !bCanCrawl {
						processedWorkItemsMutex.Lock()
						defer processedWorkItemsMutex.Unlock()
						processedWorkItems++
						a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
						return
					}

					// Let's read the body
					header, baseNode, err := a.crawler.Crawl(ctx, url)
					if err != nil {
						// TODO [LP]: We should at least increment a metric here.
						processedWorkItemsMutex.Lock()
						defer processedWorkItemsMutex.Unlock()
						processedWorkItems++
						a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
						return
					}

					// Let's input this through the post connection filter chain
					output := a.postCrawlProcessDomainProperties(&domainProperties, header, baseNode)
					if output == nil {
						processedWorkItemsMutex.Lock()
						defer processedWorkItemsMutex.Unlock()
						processedWorkItems++
						a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
						return
					}

					jsonString, err := output.ToJSONString()
					if err != nil {
						zap.L().Fatal("Unable to unmarshall JSON", zap.String("err", err.Error()))
					}
					fileOutputBuffer.AppendToFile(jsonString)

					appendedToFileMutex.Lock()
					if !bHasAppendedToFileOnce {
						bHasAppendedToFileOnce = true
					}
					appendedToFileMutex.Unlock()

				}(ctx, domainProperties)

			}

			lineIndex++
		}

		// Let's wait until all connectors are complete
		for i := 0; i < a.config.GeneralOptions.ConcurrentConnections; i++ {
			<-availableConnectors
		}

		// Force output at this stage, we don't have any more lines to read.
		fileOutputBuffer.Flush()

		// Save task state
		taskState.BIsFinished = true
		a.db.SavePostCrawlFilterTaskState(zoneName, taskState)

	}

	if !bHasAppendedToFileOnce {
		// Application needs to enter to error state as there won't be any post-crawl filter
		// cache file for indexing to continue
		a.applicationStateManager.OnErrored(
			"Post-crawl filters didn't produce any output. Unable to continue indexing.",
		)
		return
	}

	a.applicationStateManager.OnPostCrawlFiltersFinished()

	zap.L().Info("Post-crawl filter tasks are finished.")

}

func (a *Application) postCrawlProcessDomainProperties(
	inProperties *common.DomainProperties, header *http.Header, baseNode *html.Node,
) *common.DomainProperties {
	// Send output of a filter to next one until finished.

	for i, postCrawlFilter := range a.postCrawlFilterArray {
		output := (*postCrawlFilter).Input(inProperties, header, baseNode)
		if output == nil {
			return nil
		}

		if i+1 < len(a.postCrawlFilterArray) {
			// Chain output to the next postCrawlFilter
			inProperties = output
		} else {
			return output
		}
	}
	return nil
}
