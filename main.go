package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wishlist-bot/chatapp"

	"github.com/bwmarrin/discordgo"
)

//Environment variables

var discordBotToken string
var slackBotToken string
var referralTag string
var devMode bool
var reportDataPath string
var htmlStoragePath string

func main() {
	//Set Environment variables

	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	referralTag = os.Getenv("AMZN_TAG")
	devMode = os.Getenv("DEV") == "TRUE"
	reportDataPath = os.Getenv("AMZN_PRODUCT_LOG_PATH")
	htmlStoragePath = os.Getenv("AMZN_PRODUCT_LOG_TMP_PATH")

	fmt.Printf(`
	Starting Bot
	Discord Bot Token: %v
	Referral Tag: %v
	Dev Mode:%v
	HTML Storage Path:%v
	Reported Products Path: %v
	`,
		discordBotToken,
		referralTag,
		devMode,
		htmlStoragePath,
		reportDataPath)

	//Bot session setup

	discordSession, _ := discordgo.New(discordBotToken)
	if err := discordSession.Open(); err != nil {
		panic(err)
	}
	defer discordSession.Close()

	slackBot := chatapp.NewSlackSession(slackBotToken)
	go slackBot.Start(":6969")

	//Amazing Bot setup

	masterFetcher := masterFetcher{
		Fetcher:              HTTPFetcher{},
		ProductStorage:       NewCacheRepo(time.Second * 5),
		MessageIDProductRepo: NewCacheRepo(time.Second * 5),
		HTMLStorage:          &fileStorage{Extension: "html"},
		ReportHandler:        onReport,
		ErrorHandler:         logError,
	}
	amazingBot := AmazingBot{
		Fetcher:            &masterFetcher,
		ProductSentHandler: masterFetcher.createProductSentHandler(),
		ReportHandler:      masterFetcher.createReportHandler(),
	}
	amazingBot.Hook(chatapp.NewDiscordSession(discordSession))
	amazingBot.Hook(slackBot)

	//Code for closing the program (Ctrl+C)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func WithPromoCode(URL url.URL, tag string) url.URL {
	query, _ := url.ParseQuery(URL.RawQuery)
	query.Set("tag", tag)
	URL.RawQuery = query.Encode()
	return URL
}

type fileStorage struct {
	Extension string
}

func (s *fileStorage) Save(id string, data []byte) error {
	fileSafeID := sha256.Sum256([]byte(id))
	fileName := fmt.Sprintf("%s/%x.%s", htmlStoragePath, fileSafeID, s.Extension)
	return ioutil.WriteFile(fileName, data, 0644)
}
func (s *fileStorage) Get(id string) ([]byte, error) {
	fileSafeID := sha256.Sum256([]byte(id))
	fileName := fmt.Sprintf("%s/%x.%s", htmlStoragePath, fileSafeID, s.Extension)
	return ioutil.ReadFile(fileName)
}

func onReport(product *chatapp.Product, html []byte) {
	testData := ProductTestData{
		Product: *product,
		HTML:    string(html),
	}

	jsonData, error := json.MarshalIndent(testData, "", "\t")

	if error != nil {
		logError(error)
		return
	}
	fileName := fmt.Sprintf("%s/%d.%s", reportDataPath, time.Now().Unix(), "json")
	fmt.Printf("Report Received: %s", fileName)
	if error = ioutil.WriteFile(fileName, jsonData, 0644); error != nil {
		logError(error)
	}
}

func logError(e error) {
	fmt.Println(e)
}

/*
OLD CODE (TODO: Remove)
-------

//onProblemReport when a user reacts with a negative emoji (reporting the reply of the bot)
func onProblemReport(s chatapp.Session, messageID string) {
	logID := getLogIDFromMessage(messageID)
	if len(logID) == 0 {
		return
	}
	fmt.Printf("Product was reported [%v]\n", logID)

	os.Rename(fmt.Sprintf(amazonLogTmpPath+".json", logID), fmt.Sprintf(amazonLogPath+".json", logID))
	os.Rename(fmt.Sprintf(amazonLogTmpPath+".html", logID), fmt.Sprintf(amazonLogPath+".html", logID))
}

//fetchProduct and send it back to the chatbot to deliver to channels/chat servers
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
*/
