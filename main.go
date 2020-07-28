package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"wishlist-bot/scrapers"
	"wishlist-bot/scrapers/amazonscraper"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

//Environment variables

var botToken string
var referralTag string
var devMode bool
var amazonLogPath string
var amazonLogTmpPath string

//More globals
//TODO: Don't use globals lol

var amazonProductCache AmazonProductRepository
var cachedUrlsToLogMap map[string]string
var messageToLogMap map[string]string

func main() {
	//Set Environment variables

	botToken = os.Getenv("BOT_TOKEN")
	referralTag = os.Getenv("AMZN_TAG")
	devMode = os.Getenv("DEV") == "TRUE"
	amazonLogPath = os.Getenv("AMZN_PRODUCT_LOG_PATH")
	amazonLogTmpPath = os.Getenv("AMZN_PRODUCT_LOG_TMP_PATH")
	fmt.Printf("Starting Bot\nBot Token: %v\nReferral Tag: %v\nDev Mode:%v\nAmazon Product Log Path:%v\n\n", botToken, referralTag, devMode, amazonLogPath)

	//Initialize Globals

	amazonProductCache = &MemoryStorageToRepositoryAdapter{NewMemoryStorage()}
	cachedUrlsToLogMap = make(map[string]string)
	messageToLogMap = make(map[string]string)

	//Discord session setup

	discordSession, _ := discordgo.New(botToken)
	if err := discordSession.Open(); err != nil {
		panic(err)
	}
	defer discordSession.Close()

	//Add amazon chatbot logic

	hookAmazonChatBotToSession(NewDiscordChatAppSession(discordSession), fetchProduct, onProblemReport)

	//Code for closing the program (Ctrl+C)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

//funcs to map messages to logged products
//TODO: Use file storage instead of memory, this can get big
func addLogID(messageID string, logID string) {
	messageToLogMap[messageID] = logID
}
func getLogIDFromMessage(messageID string) string {
	return messageToLogMap[messageID]
}

func onProblemReport(cas ChatAppSession, messageID string) {
	logID := getLogIDFromMessage(messageID)
	if len(logID) == 0 {
		return
	}
	fmt.Printf("Product was reported [%v]\n", logID)

	os.Rename(fmt.Sprintf(amazonLogTmpPath+".json", logID), fmt.Sprintf(amazonLogPath+".json", logID))
	os.Rename(fmt.Sprintf(amazonLogTmpPath+".html", logID), fmt.Sprintf(amazonLogPath+".html", logID))
}

func fetchProduct(link string, send SendProduct) {
	if devMode {
		PrintMemUsage()
	}

	url, error := url.Parse(link)
	if error != nil {
		return
	}
	cacheID := fmt.Sprintf("%v://%v%v", url.Scheme, url.Host, url.Path)
	fmt.Printf("Normalized Product URL: %v\n", cacheID)

	cachedProduct, error := amazonProductCache.Get(cacheID)
	if cachedProduct != nil && error == nil {
		newMessageID := send(cachedProduct)
		messageToLogMap[newMessageID] = cachedUrlsToLogMap[cacheID]
		return
	}

	amazonScraper := amazonscraper.NewSimpleProductScraperRoutine(func(p amazonscraper.OnProductParams) {

		p.Product.URL = amazonscraper.WithPromoCode(p.Product.URL, referralTag)
		newMessageID := send(p.Product)

		amazonProductCache.Save(cacheID, p.Product)

		newLogID := uuid.New().String()
		cachedUrlsToLogMap[cacheID] = newLogID
		messageToLogMap[newMessageID] = newLogID

		tmpSavePath := fmt.Sprintf(amazonLogTmpPath, newLogID)
		amazonscraper.CreateTestDataOnProduct(tmpSavePath, p)
	})

	startingRequest, _ := http.NewRequest("GET", link, nil)
	startingRequest.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.122 Safari/537.36")
	go amazonScraper.Run(scrapers.HTTPStepParameters{
		Request: startingRequest,
	})
}
