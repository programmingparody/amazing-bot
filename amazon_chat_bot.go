package main

import (
	"net/url"
	"wishlist-bot/scrapers/amazonscraper"
)

// FetchProductFunc is a call signature that takes a link and callbacks with product params
type FetchProductFunc func(link string, send SendProduct)
type SendProduct func(*amazonscraper.Product) (messageID string)

func hookAmazonChatBotToSession(cas ChatAppSession, fp FetchProductFunc, op OnProductProblemReportCallback) {
	cas.OnMessage(createMessageHandler(fp))
	cas.OnProductProblemReport(op)
}

func createMessageHandler(fp FetchProductFunc) func(c ChatAppSession, m *Message) {
	return func(c ChatAppSession, m *Message) {
		handleMessage(c, m, fp)
	}
}

func handleMessage(c ChatAppSession, m *Message, fetchProduct FetchProductFunc) {
	// Ignore all messages created by the bot
	if m.MessageIsFromThisBot {
		return
	}

	amazonLinks := amazonscraper.ExtractManyAmazonProductLinkFromString(m.Content)

	if len(amazonLinks) == 0 {
		return
	}
	for _, link := range amazonLinks {
		fetchProduct(link, func(p *amazonscraper.Product) string {
			return onProduct(c, m, p)
		})
	}
}

func onProduct(c ChatAppSession, m *Message, p *amazonscraper.Product) string {
	_, wholeMessageAsURLError := url.Parse(m.Content)
	if wholeMessageAsURLError == nil {
		go m.Remove()
	}
	id, _ := m.RespondWithAmazonProduct(p)
	return id
}
