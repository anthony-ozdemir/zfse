package indexers

import (
	"math/rand"
	"time"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type RandomIndexer struct {
	outputScoreMap map[string]float64
	outputLimit    int64
}

func (i *RandomIndexer) Initialize(config config.TaskHandlerOptions, baseFolder string, outputLimit int64) error {
	// Read config
	i.outputScoreMap = make(map[string]float64)
	i.outputLimit = outputLimit

	return nil
}

func (i *RandomIndexer) Index(id string, properties common.DomainProperties) error {
	if i.outputLimit <= 0 {
		return nil
	}

	if len(i.outputScoreMap) > int(i.outputLimit) {
		return nil
	}

	rand.Seed(time.Now().UnixNano())

	i.outputScoreMap[id] = rand.Float64()

	return nil
}

func (i *RandomIndexer) Query(userQuery string) (map[string]float64, error) {
	return i.outputScoreMap, nil
}

func (i *RandomIndexer) GetType() string {
	return "builtin.random_indexer"
}
