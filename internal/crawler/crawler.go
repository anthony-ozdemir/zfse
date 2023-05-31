package crawler

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jimsmart/grobotstxt"
	"golang.org/x/net/html"
)

const agentName = "zfse/1.0"
const userAgent = "Mozilla/5.0 (compatible; zfse/1.0; +https://www.zfse.org)"

type CrawlerOptions struct {
	TimeOutInSeconds        int
	MinContentLengthInBytes int64
	MaxContentLengthInBytes int64
	ContentReadLimitInBytes int64
}

type Crawler struct {
	opts       CrawlerOptions
	httpClient *http.Client
}

func NewCrawler(opts CrawlerOptions) *Crawler {
	c := Crawler{}
	c.opts = opts

	c.httpClient = &http.Client{
		Timeout: time.Second * time.Duration(opts.TimeOutInSeconds),
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//	return http.ErrUseLastResponse
		// },
	}

	return &c
}

func (c *Crawler) setHeaders(request *http.Request) {
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
}

func (c *Crawler) HasDNSARecord(ctx context.Context, domainName string) bool {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(c.opts.TimeOutInSeconds)*time.Second)
	defer cancel()

	_, err := net.DefaultResolver.LookupIP(ctxTimeout, "ip", domainName)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (c *Crawler) CanCrawl(ctx context.Context, urlString string) bool {
	parsed, err := url.Parse(urlString)
	if err != nil {
		return false
	}
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsed.Scheme, parsed.Host)
	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return false
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	ok := grobotstxt.AgentAllowed(string(content), agentName, urlString)

	return ok
}

func (c *Crawler) Crawl(ctx context.Context, urlString string) (*http.Header, *html.Node, error) {
	// Use the custom client to send the http.Head request
	respHead, err := c.httpClient.Head(urlString)
	if err != nil {
		return nil, nil, err
	}
	defer respHead.Body.Close()

	contentLengthStr := respHead.Header.Get("Content-Length")
	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing Content-Length: %s", err.Error())
	}

	if contentLength < c.opts.MinContentLengthInBytes || contentLength > c.opts.MaxContentLengthInBytes {
		return nil, nil, fmt.Errorf("Content-Length is not suitable: %v", contentLength)
	}

	// We can continue reading the body up to the contentLength
	req, err := http.NewRequestWithContext(ctx, "GET", urlString, nil)
	if err != nil {
		return nil, nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Wrap the response body in a LimitedReader to read only the specified number of bytes.
	limitedBody := io.LimitReader(resp.Body, c.opts.ContentReadLimitInBytes)

	doc, err := html.Parse(limitedBody)
	if err != nil {
		return nil, nil, err
	}

	return &resp.Header, doc, nil
}
