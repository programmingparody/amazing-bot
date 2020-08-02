package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"wishlist-bot/scrapers"
	"wishlist-bot/scrapers/amazonscraper"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	URL  string `json:"url"`
	Type string `json:"type"`
	Auth string `json:"auth"`
}

type fetchCallback func(amazonscraper.OnProductParams)

func HandleRequest(ctx context.Context, e Event) (string, error) {
	auth := os.Getenv("AUTH_KEY")
	if auth != e.Auth {
		return "", fmt.Errorf("Permission Denied")
	}

	c := make(chan amazonscraper.OnProductParams)

	go fetchProduct(e.URL, c)
	p := <-c

	data, error := json.Marshal(struct {
		RawHTML string
		Product amazonscraper.Product
	}{
		RawHTML: string(p.RawHTML),
		Product: *p.Product,
	})
	return string(data), error
}

func fetchProduct(link string, c chan amazonscraper.OnProductParams) {
	amazonScraper := amazonscraper.NewSimpleProductScraperRoutine(func(p amazonscraper.OnProductParams) {
		c <- p
	})

	startingRequest, _ := http.NewRequest("GET", link, nil)
	startingRequest.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	amazonScraper.Run(scrapers.HTTPStepParameters{
		Request: startingRequest,
	})
}

func main() {
	lambda.Start(HandleRequest)
}
