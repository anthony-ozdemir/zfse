package crawler

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestIncompleteHTML(t *testing.T) {
	// Just to confirm that the `html.Parse` function can handle invalid or incomplete HTML documents.
	invalidHTMLDocument := `<!DOCTYPE html>
							<html>
								<head>
									<title>Example</title>
								</head>
								<body>
									<h1>Heading</h1>
									<p>This is an incomplete document`
	invalidHTMLDocument = strings.Replace(invalidHTMLDocument, "\t", "", -1)
	invalidHTMLDocument = strings.Replace(invalidHTMLDocument, "\n", "", -1)

	// Attempt to parse the invalid/incomplete HTML document
	reader := strings.NewReader(invalidHTMLDocument)
	parsedDoc, err := html.Parse(reader)
	if err != nil {
		t.Fatalf("html.Parse failed on the invalid HTML document: %v", err)
	}

	// Check if the parsedDoc has the <!DOCTYPE html> as the first child
	if parsedDoc.FirstChild.Type != html.DoctypeNode || parsedDoc.FirstChild.Data != "html" {
		t.Error("The first child of the parsed document is not the <!DOCTYPE html> tag")
	}
	titleElement := parsedDoc.FirstChild.NextSibling.FirstChild.FirstChild

	// Check that title data is correct
	if titleElement.Type != html.ElementNode || titleElement.FirstChild.Data != "Example" {
		t.Error("title element data is not correct")
	}
}

/*
// Skipping this test for now. Test requires internet connection.
// TODO [LP]: We should find a way to mock tests that require internet connection instead.
func TestKnownWebsite(t *testing.T) {
	knownWebsiteURL := "https://galaxiesofeden.com"
	opts := CrawlerOptions{
		TimeOutInSeconds:        5,
		MinContentLengthInBytes: 1024,
		MaxContentLengthInBytes: 102400,
		ContentReadLimitInBytes: 1024,
	}
	crawler := NewCrawler(opts)
	bCanCrawl := crawler.CanCrawl(context.Background(), knownWebsiteURL)
	assert.True(t, bCanCrawl)
	header, baseNode, err := crawler.Crawl(context.Background(), knownWebsiteURL)
	require.NoError(t, err)

	assert.NotNil(t, header)
	assert.NotNil(t, baseNode)

}
*/
