package amazonscraper

import (
	"bytes"
	"encoding/json"
	"strings"
	"wishlist-bot/scrapers"

	"github.com/PuerkitoBio/goquery"
)

//Product represents a scraped Amazon.com product
type Product struct {
	Title         string
	Price         float32
	ImageURL      string
	Description   string
	RatingsCount  uint    //Number of ratings (the amound of people giving a product a star count)
	Rating        float32 //Rating percentage (0-5, 5 = five star rating, 4.5 would be 4 and half stars)
	OutOfStock    bool
	OriginalPrice float32
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

//ParseProductHTML (html) into a *Product
func ParseProductHTML(html []byte) (*Product, error) {

	document, error := goquery.NewDocumentFromReader(bytes.NewBuffer(html))
	if error != nil {
		return nil, error
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
	return &Product{
		Title:         strings.Trim(productTitle, "\n "),
		Price:         float32(price),
		ImageURL:      productImageURL,
		Description:   strings.Trim(productDescription.Find("p").Text(), "\n "),
		RatingsCount:  uint(numbersFromStringFallback(ratingsCountText, 0)[0]),
		Rating:        float32(numbersFromStringFallback(ratingsText, 0)[0]),
		OutOfStock:    outOfStock,
		OriginalPrice: originalPrice,
	}, nil
}
