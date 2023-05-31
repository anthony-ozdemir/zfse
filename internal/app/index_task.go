package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/enum"
	"github.com/anthony-ozdemir/zfse/internal/helper"
	"github.com/anthony-ozdemir/zfse/internal/path_manager"
)

const (
	totalLinesUntilIndexerDBSave = 1024
)

func createIndexID(tldName string, lineIndex int) string {
	indexID := fmt.Sprintf("%s_%d", tldName, lineIndex)
	return indexID
}

func parseIndexID(indexID string) (string, int, error) {
	lastUnderscore := strings.LastIndex(indexID, "_")
	if lastUnderscore == -1 {
		return "", 0, fmt.Errorf("invalid indexID format")
	}

	tldName := indexID[:lastUnderscore]
	lineIndexStr := indexID[lastUnderscore+1:]
	lineIndex, err := strconv.Atoi(lineIndexStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse lineIndex: %v", err)
	}

	return tldName, lineIndex, nil
}

func normalizeIndexerScore(scores map[string]float64) map[string]float64 {
	// 1. Find minimum and maximum score values in the map
	minScore, maxScore := findMinMaxIndexerScore(scores)

	// 2. Calculate the range of scores
	scoreRange := maxScore - minScore

	// 3. Iterate through the map, normalize the scores, and copy to a new map
	normalizedScores := make(map[string]float64, len(scores))

	for key, score := range scores {
		if scoreRange == 0 {
			normalizedScores[key] = 1
		} else {
			normalizedScores[key] = (score - minScore) / scoreRange
		}
	}

	return normalizedScores
}

func findMinMaxIndexerScore(scores map[string]float64) (float64, float64) {
	minScore := 0.0
	maxScore := 0.0

	first := true
	for _, score := range scores {
		if first {
			minScore = score
			maxScore = score
			first = false
		} else {
			if score > maxScore {
				maxScore = score
			}
			if score < minScore {
				minScore = score
			}
		}
	}

	return minScore, maxScore
}

func (a *Application) getTotalPostCrawlFileLinesToRead() (map[string]int, error) {
	// TODO [HP]: Performance benchmark with a +30GB file. If performance is too bad then
	// we need to estimate via read bytes. Though, this would increase code complexity.
	totalLinesMap := make(map[string]int)
	for zoneName, _ := range a.zoneFileRegistry {
		postCrawlCacheFile := path_manager.GetPostCrawlFilterOutputFilePath(zoneName)
		lines, err := helper.CountLinesInFile(postCrawlCacheFile)
		if err != nil {
			return nil, err
		}
		totalLinesMap[zoneName] = lines
	}
	return totalLinesMap, nil
}

func (a *Application) runIndexer(wg *sync.WaitGroup) {
	defer wg.Done()

	// Let's first setup work item estimates
	totalLinesMap, err := a.getTotalPostCrawlFileLinesToRead()
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

	for zoneName, _ := range a.zoneFileRegistry {
		indexerTaskState := a.db.GetIndexerTaskState(zoneName)
		if indexerTaskState.BIsFinished {
			processedWorkItems += totalLinesMap[zoneName]
			a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
			continue
		}

		postCrawlCacheFile := path_manager.GetPostCrawlFilterOutputFilePath(zoneName)
		// Let's read this file line by line and input to indexer
		file, err := os.Open(postCrawlCacheFile)
		if err != nil {
			zap.L().Fatal("Error opening file.", zap.String("err", err.Error()))
		}
		defer file.Close()

		// Create a bufio.Scanner to read the file line by line
		reader := bufio.NewReader(file)

		startIndex := indexerTaskState.LineIndex
		lineIndex := 0
		remainingLinesToNextDBSave := totalLinesUntilIndexerDBSave
		for {
			// Check if we need to gracefully shut-down before finishing this task.
			currentState := a.applicationStateManager.GetApplicationState()
			if currentState.task == enum.Shutdown {
				return
			}

			line, err := reader.ReadString('\n')

			if err != nil {
				if err == io.EOF {
					break
				}
				zap.L().Fatal("Error reading a line from file.", zap.String("err", err.Error()))
			}

			if lineIndex < startIndex {
				processedWorkItems++
				a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)
				continue
			}

			// All lines are supposed to be JSON objects at this stage
			domainProperties := common.DomainProperties{}
			err = json.Unmarshal([]byte(line), &domainProperties)
			if err != nil {
				zap.L().Fatal("Unable to parse JSON.", zap.String("err", err.Error()))
			}

			indexID := createIndexID(zoneName, lineIndex)
			err = (*a.indexer).Index(indexID, domainProperties)
			if err != nil {
				zap.L().Fatal("Unable to index.", zap.String("err", err.Error()))
			}

			remainingLinesToNextDBSave--
			lineIndex++

			processedWorkItems++
			a.applicationStateManager.SetProcessedWorkItems(processedWorkItems)

			if remainingLinesToNextDBSave <= 0 {
				remainingLinesToNextDBSave = totalLinesUntilIndexerDBSave
				indexerTaskState.LineIndex = lineIndex
				a.db.SaveIndexerTaskState(zoneName, indexerTaskState)
			}
		}

		indexerTaskState.BIsFinished = true
		a.db.SaveIndexerTaskState(zoneName, indexerTaskState)
	}

	a.applicationStateManager.OnReadyToSearch()

	zap.L().Info("Indexer tasks are finished. Ready to search!")

}

func (a *Application) queryIndexer(userQuery string) []common.DomainProperties {
	output, err := (*a.indexer).Query(userQuery)
	if err != nil {
		zap.L().Fatal(
			"Unable to index.", zap.String("indexer_type", (*a.indexer).GetType()),
			zap.String("err", err.Error()),
		)
	}

	// Let's normalize the output scores between 0.0 and 1.0
	normalizedScores := normalizeIndexerScore(output)

	// Let's sort the output
	type IDScore struct {
		ID    string
		Score float64
	}

	idScores := make([]IDScore, 0, len(output))
	// Populate the slice with values from the output map
	for id, score := range output {
		idScores = append(idScores, IDScore{ID: id, Score: score})
	}

	// Sort the slice based on scores, in descending order
	sort.Slice(
		idScores, func(i, j int) bool {
			return idScores[i].Score > idScores[j].Score
		},
	)

	sortedDomainProperties := make([]common.DomainProperties, 0)
	for _, idScore := range idScores {
		tldName, lineIndex, err := parseIndexID(idScore.ID)
		if err != nil {
			zap.L().Fatal("Invalid indexID.", zap.String("err", err.Error()))
		}

		// TODO [HP]: This is super inefficient. It looks like we need to store lineOffsetBytes in another file then
		// use the known offsets to read in one-shot. Though, this might create memory issues too. It will take some
		// time to implement this efficiently.
		postCrawlCacheFile := path_manager.GetPostCrawlFilterOutputFilePath(tldName)
		line, err := helper.ReadLine(postCrawlCacheFile, lineIndex)
		if err != nil {
			zap.L().Fatal("Unable to read line.", zap.String("err", err.Error()))
		}

		// All lines are supposed to be JSON objects at this stage
		domainProperties := common.DomainProperties{}
		err = json.Unmarshal([]byte(line), &domainProperties)
		if err != nil {
			zap.L().Fatal("Unable to parse JSON.", zap.String("err", err.Error()))
		}

		// Let's also record the index score
		domainProperties.FloatProperties["indexer_score"] = idScore.Score
		domainProperties.FloatProperties["normalized_indexer_score"] = normalizedScores[idScore.ID]

		sortedDomainProperties = append(sortedDomainProperties, domainProperties)

	}

	return sortedDomainProperties
}
