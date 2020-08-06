package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/programmingparody/amazing-bot/chatapp"
	"github.com/programmingparody/amazing-bot/scrapers/amazonscraper"
)

//HTTPFetcher sends a request to Amazon, parses the HTML, and returns a Product
type HTTPFetcher struct {
	Cookies []http.Cookie
}

//Fetch Product from URL
func (hf *HTTPFetcher) Fetch(url *url.URL) (*chatapp.Product, error) {
	html, error := hf.GetHTML(url)
	if error != nil {
		return nil, error
	}
	amazonProduct, error := amazonscraper.ParseProductHTML(html)
	if error != nil {
		return nil, error
	}

	product := amazonToChatAppProduct(url, amazonProduct)
	return &product, nil
}

//GetHTML data from the URL parameter
func (hf *HTTPFetcher) GetHTML(url *url.URL) ([]byte, error) {
	startingRequest, _ := http.NewRequest("GET", url.String(), nil)
	startingRequest.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")

	for _, c := range hf.Cookies {
		startingRequest.AddCookie(&c)
	}
	response, error := http.DefaultClient.Do(startingRequest)

	if error != nil {
		return nil, error
	}
	return ioutil.ReadAll(response.Body)
}
