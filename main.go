package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wishlist-bot/scrapers"
	"wishlist-bot/scrapers/amazonscraper"

	"github.com/bwmarrin/discordgo"
)

var botToken string
var referralTag string
var devMode bool
var amazonLogPath string

//Keep track of messages that we've responded to
//Used for when a problem reaction comes in so we can create a test object for it
var messageResponseLogs *MemoryStorage

var amazonProductCache AmazonProductRepository

func main() {
	//Environment variables
	botToken = os.Getenv("BOT_TOKEN")
	referralTag = os.Getenv("AMZN_TAG")
	devMode = os.Getenv("DEV") == "TRUE"
	amazonLogPath = os.Getenv("AMZN_PRODUCT_LOG_PATH")
	fmt.Printf("Starting Bot\nBot Token: %v\nReferral Tag: %v\nDev Mode:%v\nAmazon Product Log Path:%v\n\n", botToken, referralTag, devMode, amazonLogPath)

	//Discord Session
	discordSession, _ := discordgo.New(botToken)
	if err := discordSession.Open(); err != nil {
		panic(err)
	}
	defer discordSession.Close()

	//Setup product repositories
	messageResponseLogs = NewMemoryStorage()
	amazonProductCache = &MemoryStorageToRepositoryAdapter{NewMemoryStorage()}

	hookAmazonChatBotToSession(NewDiscordChatAppSession(discordSession), fetchProduct, onProblemReport)

	//Code for closing the program (Ctrl+C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func logProductParams(p amazonscraper.OnProductParams) {
	amazonscraper.CreateTestDataOnProduct(fmt.Sprintf(amazonLogPath, time.Now().String()), p)
}

func onProblemReport(cas ChatAppSession, messageID string) {
	data, error := messageResponseLogs.Get(messageID)
	if error != nil {
		return
	}
	params := data.(amazonscraper.OnProductParams)
	go logProductParams(params)
}

func fetchProduct(link string, send SendProduct) {
	url, error := url.Parse(link)
	if error != nil {
		return
	}
	cacheID := fmt.Sprintf("%v://%v%v", url.Scheme, url.Host, url.Path)
	fmt.Printf("Normalized Product: %v\n", cacheID)

	cachedProduct, error := amazonProductCache.Get(cacheID)
	if cachedProduct != nil && error == nil {
		fmt.Printf("Sending Cached Product\n")
		send(cachedProduct)
		return
	}

	amazonScraper := amazonscraper.NewSimpleProductScraperRoutine(func(p amazonscraper.OnProductParams) {
		fmt.Printf("Sending Fetched Product\n")
		p.Product.URL = amazonscraper.WithPromoCode(p.Product.URL, referralTag)
		newMessageID := send(p.Product)
		messageResponseLogs.Save(newMessageID, p)
		amazonProductCache.Save(cacheID, p.Product)

		//In Dev mode we save every product we parse
		if devMode {
			PrintMemUsage()
			go logProductParams(p)
		}
	})

	startingRequest, _ := http.NewRequest("GET", link, nil)
	startingRequest.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	go amazonScraper.Run(scrapers.HTTPStepParameters{
		Request: startingRequest,
	})
}
