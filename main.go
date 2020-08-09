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

	"github.com/programmingparody/amazing-bot/chatapp"

	"github.com/bwmarrin/discordgo"
)

//Environment variables

var discordBotToken string
var slackBotToken string
var amazonReferralTag string
var devMode bool
var reportDataPath string
var htmlStoragePath string
var slackWebPort string

func main() {
	config := readConfigFromFile("./config.json")
	//Set Environment variables

	discordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	slackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	amazonReferralTag = os.Getenv("AMZN_REFERRAL_TAG")
	devMode = os.Getenv("DEV") == "TRUE"
	reportDataPath = os.Getenv("REPORT_PATH")
	htmlStoragePath = os.Getenv("HTML_STORAGE_PATH")
	slackWebPort = os.Getenv("SLACK_WEB_PORT")

	fmt.Printf(`
	========================
	Amazing Chat Bot Started
	------------------------
	Discord Bot Token: %v
	Amazon Referral Tag: %v
	Dev Mode:%v
	HTML Storage Path:%v
	Reported Products Path: %v
	========================

	`,
		discordBotToken,
		amazonReferralTag,
		devMode,
		htmlStoragePath,
		reportDataPath)

	//Bot session setup

	discordSession, _ := discordgo.New(discordBotToken)
	if err := discordSession.Open(); err != nil {
		panic(err)
	}
	defer discordSession.Close()

	slackBot := chatapp.NewSlackSession(slackBotToken, "-1")
	go slackBot.Start(slackWebPort)

	//Amazing Bot setup

	masterFetcher := masterFetcher{
		Fetcher:              HTTPFetcher{Cookies: config.HTTPCookies},
		ProductStorage:       newCacheRepo(time.Second * 5),
		MessageIDProductRepo: newCacheRepo(time.Second * 5),
		HTMLStorage:          &fileStorage{Extension: "html"},
		ReportHandler:        onReport,
		ErrorHandler:         logError,
		ProductModifier: func(p *chatapp.Product) {
			applyPromoCode(p.URL, amazonReferralTag)
		},
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

func applyPromoCode(URL *url.URL, tag string) {
	query, _ := url.ParseQuery(URL.RawQuery)
	query.Set("tag", tag)
	URL.RawQuery = query.Encode()
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
