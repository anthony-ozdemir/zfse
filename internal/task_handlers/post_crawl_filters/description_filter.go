package post_crawl_filters

import (
	"net/http"
	"regexp"

	"go.uber.org/zap"
	"golang.org/x/net/html"

	"github.com/anthony-ozdemir/zfse/internal/common"
	"github.com/anthony-ozdemir/zfse/internal/config"
)

type DescriptionFilter struct {
	// Config
	descriptionRegex regexp.Regexp
	outputArray      []common.DomainProperties
}

func (d *DescriptionFilter) Initialize(config config.TaskHandlerOptions) error {
	// Create maps
	d.outputArray = make([]common.DomainProperties, 0)

	descriptionRegexString, ok := config.StringOptions["description_regex"]
	if !ok {
		zap.L().Fatal("Unable to find description_regex config option.")
	}

	descriptionRegex, err := regexp.Compile(descriptionRegexString)
	if err != nil {
		zap.L().Fatal("Unable to compile regex.", zap.String("err", err.Error()))
	}

	d.descriptionRegex = *descriptionRegex

	return nil
}

func (d *DescriptionFilter) Input(
	inProperties *common.DomainProperties, header *http.Header,
	baseNode *html.Node,
) *common.DomainProperties {

	description := ""
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "meta" {
			var nameAttr, contentAttr *html.Attribute
			for _, attr := range node.Attr {
				attrCopy := attr
				if attrCopy.Key == "name" && attrCopy.Val == "description" {
					nameAttr = &attrCopy
				}
				if attrCopy.Key == "content" {
					contentAttr = &attrCopy
				}
			}

			if nameAttr != nil && contentAttr != nil {
				description = contentAttr.Val
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}

	crawler(baseNode)

	if d.descriptionRegex.MatchString(description) {
		inProperties.StringProperties["description"] = description
		return inProperties
	} else {
		return nil
	}
}

func (d *DescriptionFilter) GetType() string {
	return "builtin.description_filter"
}
