package amazonscraper

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"wishlist-bot/scrapers"
)

type CreateSession struct {
	Cookies map[string]string
}

func (cs *CreateSession) Step(params scrapers.HTTPStepParameters) (scrapers.HTTPStepParameters, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	url, _ := url.Parse("https://www.amazon.com")

	for key, value := range cs.Cookies {
		jar.SetCookies(url, []*http.Cookie{
			{
				Name:  key,
				Value: value,
			},
		})
	}

	client := http.Client{
		Jar: jar,
	}
	params.Client = &client
	return params, nil
}
