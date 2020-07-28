package amazonscraper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
	"wishlist-bot/scrapers"
)

func testFile(file string, t *testing.T) {
	testData := ReadFileToProductTestData(file)

	testScraperRoutine := scrapers.NewEmtpyHTTPRoutine()
	testScraperRoutine.AddHTTPStep(&ParseProduct{
		OnProduct: func(p OnProductParams) {
			if reflect.DeepEqual(testData.Product, *p.Product) {
				return
			}
			t.Errorf("Not equal:\nv==============Expected==============v\n%v\n\n\nv==============Result==============v\n%v", testData.Product, *p.Product)
		},
	})

	htmlString := []byte(testData.HTML)

	mockResponse := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(htmlString)),
	}

	testScraperRoutine.Run(scrapers.HTTPStepParameters{
		Response: &mockResponse,
		Request: &http.Request{
			URL: &testData.Product.URL,
		},
	})

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
