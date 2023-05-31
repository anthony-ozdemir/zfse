package indexers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

func TestBasicIndexer(t *testing.T) {
	// Test setup
	conf := config.TaskHandlerOptions{
		Type:          "builtin.basic_indexer",
		StringOptions: make(map[string]string),
		IntOptions:    make(map[string]int64),
		FloatOptions:  make(map[string]float64),
		BoolOptions:   make(map[string]bool),
	}

	indexer := BasicIndexer{}

	// Create a temporary folder for the index database
	tempFolderPath, err := os.MkdirTemp("", "indexer_temp_folder")
	require.NoError(t, err)

	// Initialize the indexer
	err = indexer.Initialize(conf, tempFolderPath, 100)
	require.NoError(t, err)

	// Create domain properties for testing
	domainproperties01 := common.NewDomainProperties()
	domainproperties01.DomainName = "example-01.com"
	domainproperties01.StringProperties["description"] = "Explore the latest electronics, gadgets, and mobile " +
		"devices at Example-01. Find the best products at affordable prices, and get top-rated customer support."

	domainproperties02 := common.NewDomainProperties()
	domainproperties02.DomainName = "example-02.com"
	domainproperties02.StringProperties["description"] = "Visit Example-02 for the latest news and updates in" +
		" technology, gaming, and entertainment. Stay up-to-date with our expert analysis and unbiased reviews."

	// Test indexing
	err = indexer.Index("example_01", domainproperties01)
	require.NoError(t, err)
	err = indexer.Index("example_02", domainproperties02)
	require.NoError(t, err)

	// Test query
	query := "GAMING NEWS. Entertainment. Test."
	scoreMap, err := indexer.Query(query)
	require.NoError(t, err)
	assert.Len(t, scoreMap, 1)
	_, ok := scoreMap["example_02"]
	assert.True(t, ok)
}
