package amazonscraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"wishlist-bot/scrapers"

	"github.com/PuerkitoBio/goquery"
)

type OnProductParams struct {
	Product        *Product
	StepParameters scrapers.HTTPStepParameters
	RawHTML        []byte
}
type ParseProduct struct {
	OnProduct func(OnProductParams)
}

type Product struct {
	Title         string  `json:"title"`
	Price         float32 `json:"price"`
	ImageURL      string  `json:"imageURL"`
	Description   string  `json:"description"`
	RatingsCount  uint    `json:"ratingsCount"` //Number of ratings (the amound of people giving a product a star count)
	Rating        float32 `json:"rating"`       //Rating percentage (0-5, 5 = five star rating, 4.5 would be 4 and half stars)
	OutOfStock    bool    `json:"outOfStock"`
	OriginalPrice float32 `json:"originalPrice"`
	URL           url.URL `json:"url"`
}

func findFallback(document *goquery.Document, selectors ...string) (selection *goquery.Selection, found bool) {
	found = false
	for _, selector := range selectors {
		selection = document.Find(selector)
		if selection.Nodes != nil {
			found = true
			break
		}
	}

	return selection, found
}

func numbersFromStringFallback(input string, defaultValue float64) []float64 {
	result := scrapers.NumbersFromString(input)
	return append(result, defaultValue)
}

func (pp *ParseProduct) Step(params scrapers.HTTPStepParameters) (scrapers.HTTPStepParameters, error) {
	//Error checks

	response := params.Response
	if response == nil {
		return params, fmt.Errorf("No Reponse")
	}
	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	document, error := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if error != nil {
		return params, error
	}

	//Parsing

	productImageElement, _ := findFallback(document, `#imgTagWrapperId img:last-child`)
	productImageURL := productImageElement.AttrOr("data-old-hires", "")
	if len(productImageURL) == 0 || productImageURL[0:5] == "data:" {
		productImageURL = productImageElement.AttrOr("src", "")

		productImageDataJSON := productImageElement.AttrOr("data-a-dynamic-image", "{}")
		productImageData := make(map[string][]int)
		json.Unmarshal([]byte(productImageDataJSON), &productImageData)

		highestPixels := 0
		for imageURL, resolution := range productImageData {
			totalImagePixels := resolution[0] * resolution[1]
			if totalImagePixels > highestPixels {
				productImageURL = strings.Replace(imageURL, "images-na.ssl-images-", "www.", 1)
				highestPixels = totalImagePixels
			}
		}
	}

	productDescription := document.Find(`#productDescription`)
	productTitle := productImageElement.AttrOr("alt", "")
	if len(productTitle) == 0 {
		productElement, _ := findFallback(document, "#productTitle")
		parentalAdvisoryElement := productElement.Find("#parentalAdvisory")
		//Frick the rules, element data may come in handly
		parentalAdvisoryElement.Remove()

		productTitle = productElement.Text()
	}

	priceElement, _ := findFallback(document, "#price_inside_buybox", "#priceblock_ourprice")
	priceText := priceElement.First().Text()
	price := numbersFromStringFallback(priceText, 0)[0]

	originalPriceElement, _ := findFallback(document, "span.priceBlockStrikePriceString.a-text-strike")
	originalPrice := float32(numbersFromStringFallback(originalPriceElement.Text(), 0)[0])

	outOfStockElement, found := findFallback(document, `#almOutOfStockAvailability_feature_div`, `#availability > span`)
	outOfStock := false
	outOfStockText := strings.Trim(outOfStockElement.Text(), " \n")
	if found && len(outOfStockText) > 0 && outOfStockText == "Currently unavailable." {
		outOfStock = true
	}

	ratingsCountText := document.Find("#acrCustomerReviewText").First().Text()
	ratingsText := document.Find("#acrPopover").First().AttrOr("title", "0")

	pp.OnProduct(OnProductParams{
		Product: &Product{
			Title:         strings.Trim(productTitle, "\n "),
			Price:         float32(price),
			ImageURL:      productImageURL,
			Description:   strings.Trim(productDescription.Find("p").Text(), "\n "),
			RatingsCount:  uint(numbersFromStringFallback(ratingsCountText, 0)[0]),
			Rating:        float32(numbersFromStringFallback(ratingsText, 0)[0]),
			OutOfStock:    outOfStock,
			OriginalPrice: originalPrice,
			URL:           *params.Request.URL,
		},
		StepParameters: params,
		RawHTML:        data,
	})

	return params, nil
}
