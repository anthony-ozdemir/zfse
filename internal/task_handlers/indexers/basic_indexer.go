package indexers

import (
	"strings"

	"github.com/blevesearch/bleve/v2"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type BasicIndexer struct {
	index      bleve.Index
	baseFolder string
}

func (b *BasicIndexer) Initialize(config config.TaskHandlerOptions, baseFolder string, outputLimit int64) error {
	// Configure default indexing mapping
	indexMapping := bleve.NewIndexMapping()
	index, err := bleve.New(baseFolder, indexMapping)

	if err != nil {
		return err
	}

	b.index = index
	b.baseFolder = baseFolder
	return nil
}

func (b *BasicIndexer) Index(id string, properties common.DomainProperties) error {

	// Let's record Description and Body fields
	record := ""
	description, ok := properties.StringProperties["description"]
	if ok {
		record += description + " - "
	}
	body, ok := properties.StringProperties["body"]
	if ok {
		record += body
	}

	if len(record) > 0 {
		err := b.index.Index(id, record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BasicIndexer) Query(userQuery string) (map[string]float64, error) {
	// We need to lower-case the query for default bleve index.
	userQuery = strings.ToLower(userQuery)

	query := bleve.NewMatchQuery(userQuery)
	search := bleve.NewSearchRequest(query)

	results, err := b.index.Search(search)
	if err != nil {
		return nil, err
	}

	idToScoreMap := make(map[string]float64)
	for _, hit := range results.Hits {
		idToScoreMap[hit.ID] = hit.Score
	}

	return idToScoreMap, nil
}

func (b *BasicIndexer) GetType() string {
	return "builtin.basic_indexer"
}
