package pre_crawl_filters

import (
	"testing"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestUniqueDomainFilter(t *testing.T) {
	// Assuming alphabetic listing

	conf := config.TaskHandlerOptions{
		Type:          "builtin.unique_domain",
		StringOptions: make(map[string]string),
		IntOptions:    make(map[string]int64),
		FloatOptions:  make(map[string]float64),
		BoolOptions:   make(map[string]bool),
	}
	conf.BoolOptions["b_nameserver_check"] = true
	conf.BoolOptions["b_discard_properties"] = true

	filter := UniqueDomainFilter{}
	err := filter.Initialize(conf)
	assert.Equal(t, err, nil)
	assert.Equal(t, conf.Type, filter.GetType())

	generateDomainProperties := func(domainName string) common.DomainProperties {
		domainProperties := common.NewDomainProperties()
		domainProperties.DomainName = domainName
		domainProperties.StringProperties["record_type"] = "ns"
		return domainProperties
	}

	inDomainPropertiesArray := make([]common.DomainProperties, 0)
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("a.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("a.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("a.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("b.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("b.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("b.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("c.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("c.com"))
	inDomainPropertiesArray = append(inDomainPropertiesArray, generateDomainProperties("c.com"))

	outPropertiesArray := make([]common.DomainProperties, 0)
	for _, domainProperties := range inDomainPropertiesArray {
		outProperties := filter.Input(&domainProperties)
		if outProperties != nil {
			outPropertiesArray = append(outPropertiesArray, *outProperties)
		}
	}

	assert.Equal(t, len(outPropertiesArray), 3)

	assert.Equal(t, outPropertiesArray[0].DomainName, "a.com")
	assert.Equal(t, outPropertiesArray[1].DomainName, "b.com")
	assert.Equal(t, outPropertiesArray[2].DomainName, "c.com")
}
