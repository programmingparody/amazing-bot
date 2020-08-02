package chatapp

import "net/url"

//Product to be sent to a chat application
type Product struct {
	Title         string   `json:"title"`
	Price         float32  `json:"price"`
	ImageURL      string   `json:"imageURL"`
	Description   string   `json:"description"`
	RatingsCount  uint     `json:"ratingsCount"` //Number of ratings (the amound of people giving a product a star count)
	Rating        float32  `json:"rating"`       //Rating percentage (0-5, 5 = five star rating, 4.5 would be 4 and half stars)
	OutOfStock    bool     `json:"outOfStock"`
	OriginalPrice float32  `json:"originalPrice"`
	URL           *url.URL `json:"url"`
}
