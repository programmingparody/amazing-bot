package main

import (
	"wishlist-bot/scrapers/amazonscraper"
)

type SetOnMessageCallback func(ChatAppSession, *Message)
type ChatAppSession interface {
	OnMessage(SetOnMessageCallback) error
}

type Message struct {
	Content                  string
	MessageIsFromThisBot     bool
	Remove                   func() error
	RespondWithAmazonProduct func(*amazonscraper.Product) error
}
