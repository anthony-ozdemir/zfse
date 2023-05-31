package app

import (
	"sort"

	"github.com/anthony-ozdemir/zfse/internal/common"
)

func (a *Application) runRankers(input []common.DomainProperties, userQuery string) []common.DomainProperties {
	rankedOutput := make([]common.DomainProperties, 0)

	type scoreItem struct {
		DomainProperties common.DomainProperties
		Score            float64
	}

	scoreList := make([]scoreItem, 0)

	for _, domainProperties := range input {
		outputScore := 0.0
		for _, ranker := range a.rankerArray {
			copyProperties := domainProperties
			rankerNormalizedWeight := a.rankerWeightMap[(*ranker).GetType()]
			// rankerOutputScore := (*ranker).Input(&copyProperties, userQuery)
			// Let's clamp the output score between 0.0 and 1.0

			outputScore += (*ranker).Input(&copyProperties, userQuery) * rankerNormalizedWeight
		}

		item := scoreItem{
			DomainProperties: domainProperties,
			Score:            outputScore,
		}
		scoreList = append(scoreList, item)
	}

	// Sort
	sort.Slice(
		scoreList, func(i, j int) bool {
			return scoreList[i].Score > scoreList[j].Score // Sorting in descending order
		},
	)

	for _, item := range scoreList {
		rankedOutput = append(rankedOutput, item.DomainProperties)
	}

	return rankedOutput
}
