package rankers

import (
	"math/rand"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type RandomRanker struct {
}

func (r *RandomRanker) Initialize(config config.TaskHandlerOptions) error {
	// Read config

	return nil
}

func (r *RandomRanker) Input(inProperties *common.DomainProperties, userQuery string) float64 {
	return rand.Float64()
}

func (r *RandomRanker) GetType() string {
	return "builtin.random_ranker"
}
