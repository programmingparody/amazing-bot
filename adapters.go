package main

import (
	"net/url"
	"wishlist-bot/chatapp"
	"wishlist-bot/scrapers/amazonscraper"
)

func amazonToChatAppProduct(url *url.URL, p *amazonscraper.Product) chatapp.Product {
	return chatapp.Product{
		Title:         p.Title,
		Price:         p.Price,
		ImageURL:      p.ImageURL,
		Description:   p.Description,
		RatingsCount:  p.RatingsCount,
		Rating:        p.Rating,
		OutOfStock:    p.OutOfStock,
		OriginalPrice: p.OriginalPrice,
		URL:           url,
	}
}
