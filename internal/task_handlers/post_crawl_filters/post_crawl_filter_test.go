package post_crawl_filters

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

func TestDescriptionFilter(t *testing.T) {
	// Test setup
	conf := config.TaskHandlerOptions{
		Type:          "builtin.description_filter",
		StringOptions: make(map[string]string),
		IntOptions:    make(map[string]int64),
		FloatOptions:  make(map[string]float64),
		BoolOptions:   make(map[string]bool),
	}
	conf.StringOptions["description_regex"] = "^.*example meta.*$"

	descriptionFilter := DescriptionFilter{}
	err := descriptionFilter.Initialize(conf)
	require.NoError(t, err)
	assert.Equal(t, conf.Type, descriptionFilter.GetType())

	htmlDocument := `<!DOCTYPE html>
					<html>
						<head>
							<meta name="description" content="This is an example meta description."/>
							<title>Example</title>
						</head>
						<body>
							<h1>Heading</h1>
							<p>This is an example document.</p>
						</body>
					</html>`

	reader := strings.NewReader(htmlDocument)
	parsedDoc, err := html.Parse(reader)
	require.NoError(t, err)

	domainproperties := common.NewDomainProperties()
	domainproperties.DomainName = "example.com"
	output := descriptionFilter.Input(&domainproperties, nil, parsedDoc)
	assert.NotNil(t, output)
	assert.Equal(t, output.DomainName, domainproperties.DomainName)
}
