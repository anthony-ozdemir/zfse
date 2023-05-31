package interfaces

import (
	"net/http"

	"golang.org/x/net/html"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type PreConnectionFilter interface {
	Initialize(config config.TaskHandlerOptions) error

	Input(inProperties *common.DomainProperties) *common.DomainProperties

	GetType() string
}

type PostConnectionFilter interface {
	Initialize(config config.TaskHandlerOptions) error

	Input(inProperties *common.DomainProperties, header *http.Header, baseNode *html.Node) *common.DomainProperties

	GetType() string
}

type Indexer interface {
	Initialize(config config.TaskHandlerOptions, baseFolder string, outputLimit int64) error

	Index(id string, properties common.DomainProperties) error

	Query(userQuery string) (map[string]float64, error)

	GetType() string
}

type Ranker interface {
	Initialize(config config.TaskHandlerOptions) error

	Input(inProperties *common.DomainProperties, userQuery string) float64

	GetType() string
}
