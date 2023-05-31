package rankers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

func TestIndexerRanker(t *testing.T) {
	// Test setup
	conf := config.TaskHandlerOptions{
		Type:          "builtin.indexer_ranker",
		StringOptions: make(map[string]string),
		IntOptions:    make(map[string]int64),
		FloatOptions:  make(map[string]float64),
		BoolOptions:   make(map[string]bool),
	}

	indexerRanker := IndexerRanker{}
	err := indexerRanker.Initialize(conf)
	require.NoError(t, err)

	assumedScore := 0.5
	domainproperties := common.NewDomainProperties()
	domainproperties.DomainName = "example-01.com"
	domainproperties.FloatProperties["normalized_indexer_score"] = assumedScore
	score := indexerRanker.Input(&domainproperties, "")
	assert.Equal(t, score, assumedScore)
}
