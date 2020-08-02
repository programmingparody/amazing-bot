/*
Tests for reported products

Usage:
	./test.sh

Flow:
	Iterates through reported products (ProductTestData .json files). For each report:
		1) Run the HTML through amazonscraper.ParseProductHTML to get a Product with the up to date function
		2) Compare new Product with the test data's Product
		3) If the results are different, the test failed

Preprocessing ProductTestData .json files:
	Running this test on it's own after a Product was reported will NOT cause the test to fail.
	The "Product" object in the .json files have to be manually edited to match what we'd expect the Product Object to look like

	Steps to Edit ProductTestData .json files:
		1) Manually review the HTML (Or for convenience, visit the URL in a browser) to see what the correct results should be
		2) Replace missing or incorrect values with the correct results
			For example, if Product.Price = 4.99, but on Amazon we see $49.99, change Product.Price = 49.99
			Do this for all incorrect values in the .json file
		3) Run ./test.sh
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"wishlist-bot/scrapers/amazonscraper"
)

func testFile(file string, t *testing.T) {
	var testData ProductTestData
	fileData, error := ioutil.ReadFile(file)
	if error != nil {
		t.Error(error)
	}
	json.Unmarshal(fileData, &testData)

	parsedAmazonProduct, error := amazonscraper.ParseProductHTML([]byte(testData.HTML))
	parsedProduct := amazonToChatAppProduct(testData.Product.URL, parsedAmazonProduct)
	if error != nil {
		t.Error(error)
	}
	if !reflect.DeepEqual(testData.Product, parsedProduct) {
		t.Errorf("Not equal:\nv==============Expected==============v\n%v\n\n\nv==============Result==============v\n%v", testData.Product, parsedProduct)
	}
}

func TestParseProduct(t *testing.T) {
	productLogDirectory := os.Getenv("REPORT_FILE_PATH")
	files, error := ioutil.ReadDir(fmt.Sprintf(productLogDirectory, ""))

	if error != nil {
		t.Error(error)
	}

	for _, file := range files {
		name := file.Name()
		if len(name) < 5 {
			continue
		}
		extension := name[len(name)-5:]
		if file.IsDir() || extension != ".json" {
			continue
		}

		filePath := fmt.Sprintf(productLogDirectory, file.Name())
		t.Log(filePath)
		testFile(filePath, t)
	}
}
