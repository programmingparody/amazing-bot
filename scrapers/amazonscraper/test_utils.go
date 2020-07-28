package amazonscraper

import (
	"encoding/json"
	"io/ioutil"
)

type ProductTestData struct {
	Product Product
	HTML    string
}

func CreateTestDataOnProduct(filePath string, p OnProductParams) {
	testObject := ProductTestData{
		Product: *p.Product,
		HTML:    string(p.RawHTML),
	}
	jsonProduct, _ := json.MarshalIndent(testObject, "", "\t")
	ioutil.WriteFile(filePath+".json", jsonProduct, 0644)
	ioutil.WriteFile(filePath+".html", p.RawHTML, 0644)
}

func ReadFileToProductTestData(filePath string) ProductTestData {
	data, _ := ioutil.ReadFile(filePath)

	var productTest ProductTestData
	json.Unmarshal(data, &productTest)

	return productTest
}
