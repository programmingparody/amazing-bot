package main

import (
	"net/url"

	"github.com/programmingparody/amazing-bot/chatapp"
	"github.com/programmingparody/amazing-bot/scrapers/amazonscraper"
)

/*AmazingBot Chat Bot for Discord / Slack (Maybe Zoom soon)
Usage:
	Hook into chatapp sessions and set up listeners/handlers for Reports and Messages
	When a message comes through, check for Amazon links
	Valid Amazon links are ran through a Fetcher
	Successful responses from Fetcher are sent back to the chat using RespondWithProduct
*/
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
