package pre_crawl_filters

import (
	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type UniqueDomainFilter struct {
	// Config
	domainSet          map[string]bool
	bNameserverCheck   bool
	bDiscardProperties bool
}

func (f *UniqueDomainFilter) Initialize(config config.TaskHandlerOptions) error {
	// Create maps
	f.domainSet = make(map[string]bool)

	// Read configuration
	bNameserverCheck, ok := config.BoolOptions["b_nameserver_check"]
	if !ok {
		zap.L().Fatal("Unable to find b_nameserver_check config option.")
	}
	f.bNameserverCheck = bNameserverCheck

	bDiscardProperties, ok := config.BoolOptions["b_discard_properties"]
	if !ok {
		zap.L().Fatal("Unable to find b_discard_properties config option.")
	}
	f.bDiscardProperties = bDiscardProperties

	return nil
}

func (f *UniqueDomainFilter) Input(inProperties *common.DomainProperties) *common.DomainProperties {
	// Check if this URL contains a name-server
	if f.bNameserverCheck && inProperties.StringProperties["record_type"] != "ns" {
		return nil
	}

	_, bFound := f.domainSet[inProperties.DomainName]
	if !bFound {
		// Design Note: icann zone files are supposed to be alphabetic
		// Thus, we can clear the domainSet map now to minimize memory usage
		if len(f.domainSet) > 1024 {
			f.domainSet = make(map[string]bool)
		}

		f.domainSet[inProperties.DomainName] = true

		if f.bDiscardProperties {
			inProperties.StringProperties = make(map[string]string)
			inProperties.IntProperties = make(map[string]int64)
			inProperties.FloatProperties = make(map[string]float64)
			inProperties.BoolProperties = make(map[string]bool)
		}

		return inProperties
	} else {
		return nil
	}
}

func (f *UniqueDomainFilter) GetType() string {
	return "builtin.unique_domain"
}
