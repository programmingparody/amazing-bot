package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type config struct {
	HTTPCookies []http.Cookie `json:"HTTPCookies"`
}

func readConfigFromFile(filePath string) config {
	data, _ := ioutil.ReadFile(filePath)

	var config config

	json.Unmarshal(data, &config)

	return config
}
