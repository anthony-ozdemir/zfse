package pre_crawl_filters

import (
	"strings"

	"go.uber.org/zap"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type LengthFilter struct {
	// Config
	minLength int64
	maxLength int64
}

func (l *LengthFilter) Initialize(config config.TaskHandlerOptions) error {
	// Read config
	minLength, ok := config.IntOptions["min_length"]
	if !ok {
		zap.L().Fatal("Unable to find min_length config option.")
	}
	l.minLength = minLength

	maxLength, ok := config.IntOptions["max_length"]
	if !ok {
		zap.L().Fatal("Unable to find max_length config option.")
	}
	l.maxLength = maxLength

	return nil
}

func (l *LengthFilter) Input(inProperties *common.DomainProperties) *common.DomainProperties {
	// Discard last . on DNS record. This is expected to be TLD name like .com
	lastDotIndex := strings.LastIndex(inProperties.DomainName, ".")
	urlWithoutLastDot := ""
	if lastDotIndex != -1 {
		urlWithoutLastDot = inProperties.DomainName[:lastDotIndex]
	}

	// Check the number of unique characters in the url name
	if len(urlWithoutLastDot) < int(l.minLength) {
		return nil
	}

	if len(urlWithoutLastDot) > int(l.maxLength) {
		return nil
	}

	return inProperties
}

func (l *LengthFilter) GetType() string {
	return "builtin.length_filter"
}
