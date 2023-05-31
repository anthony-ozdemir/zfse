package rankers

import (
	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type IndexerRanker struct {
}

func (r *IndexerRanker) Initialize(config config.TaskHandlerOptions) error {
	// Read config

	return nil
}

func (r *IndexerRanker) Input(inProperties *common.DomainProperties, userQuery string) float64 {
	indexScore, ok := inProperties.FloatProperties["normalized_indexer_score"]
	if !ok {
		return 0
	}
	return indexScore
}

func (r *IndexerRanker) GetType() string {
	return "builtin.indexer_ranker"
}
