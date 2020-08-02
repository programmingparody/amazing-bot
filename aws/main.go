/*
AWS Lambda Function
Takes a url, fetches the product from amazon using the scraper, and returns amazonscraper.OnProductParams JSON encoded as the response

WARNING:
	Amazon.com sometimes detects that it's being hit and returns an error, use with caution

	Sample Error Response:
		To discuss automated access to Amazon data please contact api-services-support@amazon.com.
		For information about migrating to our APIs refer to our Marketplace APIs at https://developer.amazonservices.com/ref=rm_5_sv, or our Product Advertising API at https://affiliate-program.amazon.com/gp/advertising/api/detail/main.html/ref=rm_5_ac for advertising use cases.

TODO:
	Fallback fix
		If a product is failed to be fetched through AWS, put the load back on the master server. This will cause a delay in response since we'll need to do a 2nd HTTP Request
	Ideal fix (Fixes the warning)
		Get an Amazon seller account and register to https://developer.amazonservices.com/
		Create a fetchProduct type'd function that uses AmazonServices API
*/
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
		Product amazonscraper.Product
		RawHTML string
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
