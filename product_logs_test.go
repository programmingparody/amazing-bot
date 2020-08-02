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
	productLogDirectory := os.Getenv("AMZN_PRODUCT_LOG_PATH")
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
