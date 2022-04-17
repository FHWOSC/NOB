package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Parser struct {
	timestamp *time.Time
	client    *http.Client
}

var httpc = getHttpClient()

func New() *Parser {
	p := new(Parser)
	p.timestamp = nil
	p.client = getHttpClient()
	return p
}

func (p *Parser) SetTimestamp(ts time.Time) {
	p.timestamp = &ts
}

func (p *Parser) GetTimestamp() (*time.Time, bool, error) {
	doc, err := getDoc("https://intern.fh-wedel.de/~splan/index.html?typ=benutzer_vz")
	if err != nil {
		return nil, false, err
	}

	// get timestamp of splan
	generatedAtStr := doc.Find("div.splan_version").First().Text()
	parts := strings.Split(generatedAtStr, "by")
	generatedAt, err := time.Parse("02.01.2006, 15:04 ", parts[0][len("  © generated at "):])
	if err != nil {
		return nil, false, err
	}

	changed := p.checkTimestamp(generatedAt)
	if changed {
		p.timestamp = &generatedAt
	}

	return &generatedAt, changed, nil
}

func (p *Parser) checkTimestamp(ts time.Time) bool {
	if p.timestamp == nil {
		p.timestamp = &ts
		return false
	}

	return ts.After(*p.timestamp)
}

func getDoc(url string) (*goquery.Document, error) {
	res, err := httpc.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func getHttpClient() *http.Client {
	addr := os.Getenv("PROXY_ADDR")
	if addr == "" {
		return http.DefaultClient
	}

	proxyURL, err := url.Parse(addr)
	if err != nil {
		return http.DefaultClient
	}

	//adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	//adding the Transport object to the http Client
	client := &http.Client{
		Transport: transport,
	}

	return client
}
