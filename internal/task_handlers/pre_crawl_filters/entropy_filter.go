package pre_crawl_filters

import (
	"strings"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type EntropyFilter struct {
	// Config
	discardRatio float64
}

func (e *EntropyFilter) Initialize(config config.TaskHandlerOptions) error {
	// Read config
	discardRatio, ok := config.FloatOptions["discard_ratio"]
	if !ok {
		zap.L().Fatal("Unable to find discard_ratio config option.")
	}
	e.discardRatio = discardRatio

	return nil
}

func (e *EntropyFilter) Input(inProperties *common.DomainProperties) *common.DomainProperties {
	// Check the number of unique characters in the domain name
	domainName := inProperties.DomainName
	urlWithoutDomainName := ""
	lastDotIndex := strings.LastIndex(domainName, ".")
	if lastDotIndex != -1 {
		urlWithoutDomainName = domainName[:lastDotIndex]
	}

	characterSet := make(map[rune]bool)
	for _, c := range urlWithoutDomainName {
		characterSet[c] = true
	}

	uniqueCharactersQty := len(characterSet)
	totalCharacterQty := len(urlWithoutDomainName)

	ratio := float64(uniqueCharactersQty) / float64(totalCharacterQty)
	if ratio >= e.discardRatio {
		return nil
	} else {
		return inProperties
	}

}

func (e *EntropyFilter) GetType() string {
	return "builtin.discard_high_entropy"
}
