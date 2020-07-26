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

func main() {
	botToken = os.Getenv("BOT_TOKEN")
	referralTag = os.Getenv("AMZN_TAG")
	devMode = len(os.Getenv("DEV")) > 0
	amazonLogPath = os.Getenv("AMZN_PRODUCT_LOG_PATH")
	fmt.Printf("Starting Bot\nBot Token: %v\nReferral Tag: %v\nDev Mode:%v\nAmazon Product Log Path:%v\n\n", botToken, referralTag, devMode, amazonLogPath)

	discordSession, _ := discordgo.New(botToken)
	if err := discordSession.Open(); err != nil {
		panic(err)
	}
	NewDiscordChatAppSession(discordSession).OnMessage(onMessage)
	defer discordSession.Close()

	//Code for closing the program (Ctrl+C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func log(e interface{}) {
	fmt.Printf("Log: %v\n", e)
}

func fetchProduct(link string, c chan amazonscraper.OnProductParams) {
	amazonScraper := amazonscraper.NewSimpleProductScraperRoutine(func(p amazonscraper.OnProductParams) {
		c <- p
	})

	startingRequest, _ := http.NewRequest("GET", link, nil)
	startingRequest.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	go amazonScraper.Run(scrapers.HTTPStepParameters{
		Request: startingRequest,
	})
}

func onMessage(c ChatAppSession, m *Message) {
	// Ignore all messages created by the bot
	if m.MessageIsFromThisBot {
		return
	}

	amazonLinks := amazonscraper.ExtractManyAmazonProductLinkFromString(m.Content)

	if len(amazonLinks) == 0 {
		return
	}
	productParamsChannel := make(chan amazonscraper.OnProductParams)
	for _, link := range amazonLinks {
		fetchProduct(link, productParamsChannel)

		p := <-productParamsChannel
		_, wholeMessageAsURLError := url.Parse(m.Content)
		if wholeMessageAsURLError == nil {
			go m.Remove()
		} else {
			log(wholeMessageAsURLError)
		}
		if devMode {
			go amazonscraper.CreateTestDataOnProduct(fmt.Sprintf(amazonLogPath, time.Now().String()))(p)
		}

		p.Product.URL = amazonscraper.WithPromoCode(p.Product.URL, referralTag)
		error := m.RespondWithAmazonProduct(p.Product)
		if error != nil {
			log(error)
		}
	}
}
