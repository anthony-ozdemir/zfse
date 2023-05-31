package app

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/database"
	"github.com/anthony-ozdemir/zfse/internal/enum"
	"github.com/anthony-ozdemir/zfse/internal/filebuf"
	"github.com/anthony-ozdemir/zfse/internal/helper"
	"github.com/anthony-ozdemir/zfse/internal/path_manager"
)

func (a *Application) getTotalZoneFileLinesToRead() (map[string]int, error) {
	// TODO [HP]: Performance benchmark with a +30GB file. If performance is too bad then
	// we need to estimate via read bytes. Though, this would increase code complexity.
	totalLinesMap := make(map[string]int)
	for zoneName, zoneFile := range a.zoneFileRegistry {
		lines, err := helper.CountLinesInFile(zoneFile)
		if err != nil {
			return nil, err
		}
		totalLinesMap[zoneName] = lines
	}
	return totalLinesMap, nil
}

func (a *Application) runPreCrawlFilters(wg *sync.WaitGroup) {
	defer wg.Done()

	// Let's first setup work item estimates
	totalLinesMap, err := a.getTotalZoneFileLinesToRead()
	if err != nil {
		zap.L().Fatal("Unable to read zone file lines.")
	}
	totalWorkItems := 0
	for _, lines := range totalLinesMap {
		totalWorkItems += lines
	}
	a.applicationStateManager.SetTotalWorkItems(totalWorkItems)
	processedWorkItems := 0
	a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)

	// Design Note:
	// 1. Read the input zone file line by line.
	// 2. Parse the zone file properties to DomainProperties struct.
	// 3. Feed DomainProperties struct through the pre-crawl filters one by one.
	// 4. Append the output DomainProperties struct to pre-crawl filter cache file.

	bHasAppendedToFileOnce := false
	for zoneName, zoneFile := range a.zoneFileRegistry {
		// Check the state of pre-crawl from DB
		taskState := a.db.GetPreCrawlFilterTaskState(zoneName)
		if taskState.BIsFinished {
			// We can skip this zone name
			bHasAppendedToFileOnce = true
			processedWorkItems += totalLinesMap[zoneName]
			a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
			continue
		}

		// Open the zone file
		file, err := os.Open(zoneFile)
		if err != nil {
			zap.L().Fatal("Error opening zone file.", zap.String("err", err.Error()))
		}
		defer file.Close()

		// Create a bufio.Scanner to read the file line by line
		reader := bufio.NewReader(file)

		startLineIndex := taskState.LineIndex
		lineIndex := 0

		// Prepare output buffer
		outputBufferOpts := filebuf.FileOutputBufferOptions{
			BulkOutputLimit: a.config.GeneralOptions.FileBulkOutputQty,
			FilePath:        path_manager.GetPreCrawlFilterOutputFilePath(zoneName),
			OnFlushCB: func(appendedLinesQty int) {
				// Design Note: This function will be called from the same thread.
				// Thus, we can just use the current lineIndex.

				// Save Task State at this point
				taskState := database.PreCrawlFilterTaskState{
					BIsFinished: false,
					LineIndex:   lineIndex,
				}
				a.db.SavePreCrawlFilterTaskState(zoneName, taskState)
			},
		}
		outputFileBuffer := filebuf.NewFileOutputBuffer(outputBufferOpts)

		for {
			// Check if we need to gracefully shut-down before finishing this task.
			currentState := a.applicationStateManager.GetApplicationState()
			if currentState.task == enum.Shutdown {
				outputFileBuffer.Flush()
				return
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				zap.L().Fatal("Error reading a line from zone file.", zap.String("err", err.Error()))
			}

			if lineIndex >= startLineIndex {
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					// Skip empty lines
					processedWorkItems++
					a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
					continue
				}

				// TLD Zone Files are structured by whitespace separators. Thus,
				// we can parse if via strings.Fields method.
				fields := strings.Fields(line)

				// Design Note: We are interested about records with at least five fields:
				// domainName, TTL, dnsRecordClass, dnsRecordType, dnsRecordData
				if len(fields) < 5 {
					processedWorkItems++
					a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
					continue
				}

				// Prepare the initial DomainProperties
				domainProperties := common.DomainProperties{
					DomainName:       fields[0],
					StringProperties: make(map[string]string),
					IntProperties:    make(map[string]int64),
					FloatProperties:  make(map[string]float64),
					BoolProperties:   make(map[string]bool),
				}

				domainProperties.StringProperties["ttl"] = fields[1]
				domainProperties.StringProperties["record_class"] = fields[2]
				domainProperties.StringProperties["record_type"] = fields[3]
				domainProperties.StringProperties["record_data"] = strings.Join(fields[4:], " ")

				// Discard last . on DNS domain name
				// For example, example.com. will be recorded as example.com
				lastDotIndex := strings.LastIndex(domainProperties.DomainName, ".")
				if lastDotIndex != -1 {
					domainNameWithoutLastDot := domainProperties.DomainName[:lastDotIndex]
					domainProperties.DomainName = domainNameWithoutLastDot
				}

				// Let's input this through the post connection filter chain
				output := a.preCrawlProcessDomainProperties(&domainProperties)
				if output == nil {
					processedWorkItems++
					a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
					continue
				}

				jsonString, err := output.ToJSONString()
				if err != nil {
					zap.L().Fatal("Unable to unmarshall JSON", zap.String("err", err.Error()))
				}
				outputFileBuffer.AppendToFile(jsonString)
				bHasAppendedToFileOnce = true
			}

			processedWorkItems++
			a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)

			lineIndex++
		}

		// Force output at this stage, we don't have any more lines to read.
		outputFileBuffer.Flush()

		// Save task state
		taskState.BIsFinished = true
		a.db.SavePreCrawlFilterTaskState(zoneName, taskState)
	}

	if !bHasAppendedToFileOnce {
		// Application needs to enter to error state as there won't be any pre-crawl filter
		// cache file for indexing to continue
		a.applicationStateManager.OnErrored("Pre-crawl filters didn't produce any output. Unable to continue indexing.")
		return
	}

	a.applicationStateManager.OnPreCrawlFiltersFinished()

	zap.L().Info("Pre-crawl filter tasks are finished.")

}

func (a *Application) preCrawlProcessDomainProperties(inProperties *common.DomainProperties) *common.DomainProperties {
	// Send output of a filter to next one until finished.

	for i, preCrawlFilter := range a.preCrawlFilterArray {
		output := (*preCrawlFilter).Input(inProperties)
		if output == nil {
			return nil
		}

		if i+1 < len(a.preCrawlFilterArray) {
			// Chain output to the next preCrawlFilter
			inProperties = output
		} else {
			return output
		}
	}
	return nil
}
