package parser

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Parser struct {
	client *http.Client
}

var httpc = getHttpClient()

func New() *Parser {
	p := new(Parser)
	p.client = getHttpClient()
	return p
}

func GetDoc(url string) (*goquery.Document, error) {
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

func Get(url string) (*bytes.Buffer, error) {
	res, err := httpc.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	var buf = bytes.NewBufferString("")
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
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
