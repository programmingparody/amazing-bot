package main

import (
	"fmt"
	"net/url"
	"wishlist-bot/chatapp"
	"wishlist-bot/scrapers/amazonscraper"
)

type byteStorage interface {
	Get(id string) ([]byte, error)
	Save(id string, data []byte) error
}

/*
Implements ProductFetcher.
Uses repos for caching.
Logs http responses to build test cases when reported
*/
type masterFetcher struct {
	Fetcher              HTTPFetcher                                 //Results of Fetcher are put in ProductStorage
	ProductStorage       ProductRepo                                 //Checked first before using Fetcher
	ReportHandler        func(product *chatapp.Product, html []byte) //Called when a product is reported
	MessageIDProductRepo ProductRepo                                 //Keeps track of products we've respond incase it's reported
	HTMLStorage          byteStorage                                 //Keeps track of HTTP body responses for logging when reported
	ErrorHandler         func(error)
}

func (m *masterFetcher) createProductSentHandler() func(e *SentProductEvent) {
	return func(e *SentProductEvent) {
		m.MessageIDProductRepo.Save(e.NewMessageID, e.Product)
	}
}

func (m *masterFetcher) createReportHandler() func(s chatapp.Session, messageID string) {
	return func(s chatapp.Session, messageID string) {
		productSent, error := m.MessageIDProductRepo.Get(messageID)
		if error != nil {
			m.ErrorHandler(error)
			if productSent == nil {
				return
			}
		}

		id := urlToID(productSent.URL)
		html, _ := m.HTMLStorage.Get(id)
		product, _ := m.ProductStorage.Get(id)

		m.ReportHandler(product, html)
	}
}

func (m *masterFetcher) Fetch(url *url.URL) (*chatapp.Product, error) {
	id := urlToID(url)

	storedProduct, error := m.ProductStorage.Get(id)
	if storedProduct != nil && error == nil {
		return storedProduct, nil
	}
	if error != nil {
		m.ErrorHandler(error)
	}

	html, error := m.Fetcher.GetHTML(url)
	if error != nil {
		m.ErrorHandler(error)
		return nil, error
	}

	go m.HTMLStorage.Save(id, html)

	amazonProduct, error := amazonscraper.ParseProductHTML(html)
	if error != nil {
		m.ErrorHandler(error)
	}
	product := amazonToChatAppProduct(url, amazonProduct)

	m.ProductStorage.Save(id, &product)

	return &product, error
}

func urlToID(url *url.URL) string {
	return fmt.Sprintf("%v://%v%v", url.Scheme, url.Host, url.Path)
}
