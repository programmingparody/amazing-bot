package amazonscraper

import (
	"net/url"
	"wishlist-bot/scrapers"
)

//NewSimpleProductScraperRoutine is a helper function to easily start an Amazon Product Scraping Routine
func NewSimpleProductScraperRoutine(onProduct func(OnProductParams)) scrapers.HTTPRoutine {
	scraperRoutine := scrapers.NewHTTPRoutine([]scrapers.RoutineStep{
		scrapers.RoutineStepFromHTTPStep(&CreateSession{
			Cookies: map[string]string{},
		}),
		scrapers.RoutineStepFromHTTPStep(&scrapers.MakeRequestStep{}),
		scrapers.RoutineStepFromHTTPStep(&ParseProduct{
			OnProduct: onProduct,
		}),
	})
	return scraperRoutine
}

func WithPromoCode(URL url.URL, tag string) url.URL {
	query, _ := url.ParseQuery(URL.RawQuery)
	query.Set("tag", tag)
	URL.RawQuery = query.Encode()
	return URL
}
