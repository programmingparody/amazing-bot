package main

import (
	"net/url"
	"wishlist-bot/chatapp"
)

//ProductFetcher represents a type that can fetch and return a Product
type ProductFetcher interface {
	Fetch(url *url.URL) (*chatapp.Product, error)
}
