package main

import (
	"net/url"
	"wishlist-bot/chatapp"
	"wishlist-bot/scrapers/amazonscraper"
)

//AmazingBot listens for Amazon URLs and reponds with a Product
type AmazingBot struct {
	Fetcher            ProductFetcher
	ProductSentHandler func(e *SentProductEvent)
	ReportHandler      chatapp.OnProductProblemReportCallback
}

//SentProductEvent will be fired to a callback when a product is sent
type SentProductEvent struct {
	ResponseToMessage *chatapp.Message
	NewMessageID      string
	Product           *chatapp.Product
}

//Hook to a chatapp.session
func (ab *AmazingBot) Hook(s chatapp.Session) {
	s.OnMessage(ab.createOnMessageHandler())
	s.OnProductProblemReport(ab.ReportHandler)
}

func (ab *AmazingBot) createOnMessageHandler() func(c chatapp.Session, m *chatapp.Message) {
	return func(c chatapp.Session, m *chatapp.Message) {
		ab.handleMessage(c, m)
	}
}
func (ab *AmazingBot) createOnReport() func(c chatapp.Session, m *chatapp.Message) {
	return func(c chatapp.Session, m *chatapp.Message) {
		ab.handleMessage(c, m)
	}
}

// FetchProductFunc is a call signature that takes a link and callbacks with product params
type FetchProductFunc func(link string, send SendProduct)
type SendProduct func(*amazonscraper.Product) (messageID string)

func (ab *AmazingBot) handleMessage(c chatapp.Session, m *chatapp.Message) {
	// Ignore all messages created by the bot
	if m.MessageIsFromThisBot {
		return
	}

	amazonLinks := amazonscraper.ExtractManyProductLinkFromString(m.Content)

	if len(amazonLinks) == 0 {
		return
	}
	for _, link := range amazonLinks {
		url, error := url.Parse(link)
		if error != nil {
			return
		}
		p, _ := ab.Fetcher.Fetch(url)
		_, wholeMessageAsURLError := url.Parse(m.Content)
		if wholeMessageAsURLError == nil {
			go m.Actions.Remove()
		}
		id, _ := m.Actions.RespondWithProduct(p)
		if ab.ProductSentHandler != nil {
			ab.ProductSentHandler(&SentProductEvent{
				ResponseToMessage: m,
				NewMessageID:      id,
				Product:           p,
			})
		}
	}
}
