package main

import (
	"wishlist-bot/scrapers/amazonscraper"
)

type OnMessageCallback func(ChatAppSession, *Message)
type OnProductProblemReportCallback func(cas ChatAppSession, messageID string)

type ChatAppSession interface {
	OnMessage(OnMessageCallback) error
	OnProductProblemReport(OnProductProblemReportCallback) error
}

type Message struct {
	ID                       string
	Content                  string
	MessageIsFromThisBot     bool
	Remove                   func() error
	RespondWithAmazonProduct func(*amazonscraper.Product) (newMessageID string, e error)
}
