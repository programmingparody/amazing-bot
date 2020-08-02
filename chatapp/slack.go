package chatapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
)

type SlackMessageActions struct {
	event *slackEvent
	slack *Slack
}

func (a *SlackMessageActions) Remove() error {
	return nil
}
func (a *SlackMessageActions) RespondWithProduct(p *Product) (string, error) {
	e := a.event
	s := a.slack

	data := slackRichTextJSONFromProduct(e.ChannelID, p)
	req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer([]byte(data)))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.token))
	req.Header.Add("Content-type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	resData, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	fmt.Println(string(resData))
	//TODO: Return new message id
	return "", nil
}

type slackEvent struct {
	Type            string `json:"type"`
	Text            string `json:"text"`
	UserID          string `json:"user"`
	ClientMessageID string `json:"client_msg_id"`
	URL             string `json:"url"`
	Blocks          []struct {
		Elements []struct {
			Elements []struct {
				Type string `json:"type"`
				Text string `json:"text"`
				Url  string `json"url"`
			} `json:"Elements"`
		} `json:"Elements"`
	} `json:"blocks"`
	ChannelID string `json:"channel"`
	TimeStamp string `json"ts"`
}

type slackEventMessageContainer struct {
	Token string     `json:"token"`
	Event slackEvent `json:"event"`
}

func ParseEventMessage(reader io.ReadCloser) (*slackEventMessageContainer, error) {
	message := slackEventMessageContainer{}
	data, error := ioutil.ReadAll(reader)
	if error == nil {
		defer reader.Close()
		error = json.Unmarshal(data, &message)
	}
	return &message, error
}

//Event type names
const (
	slackeventMessage = "message"
)

type slackEventHandlerFunc func(e *slackEventMessageContainer, w http.ResponseWriter, r *http.Request)

//Slack Session
type Slack struct {
	typeToHandler map[string][]slackEventHandlerFunc
	token         string
}

//NewSlackSession returns a Slack session that implements chatapp.Session
func NewSlackSession(token string) *Slack {
	handlers := make(map[string][]slackEventHandlerFunc)
	handlers[slackeventMessage] = []slackEventHandlerFunc{}
	return &Slack{
		typeToHandler: handlers,
		token:         token,
	}
}

func slackRichTextJSONFromProduct(channelId string, p *Product) string {
	funcMap := template.FuncMap{
		"url": func(p *Product) string {
			return p.URL.String()
		},
		"escape": func(s string) string {
			return strings.ReplaceAll(s, `"`, `\"`)
		},
		"price": func(p *Product) string {
			if p.OriginalPrice > 0 {
				savings := p.OriginalPrice - p.Price
				percentOff := savings / p.OriginalPrice * 100
				return fmt.Sprintf("~%0.2f~\n*%0.2f*\n_%.2f (%.0f%%) off_", p.OriginalPrice, p.Price, savings, percentOff)
			}
			return fmt.Sprintf("%.2f", p.Price)
		},
		"rating": func(p *Product) string {
			return fmt.Sprintf("%.1f", p.Rating)
		},
		"cutoff": func(input string) string {
			const max = 150
			const replacement = "..."
			if len(input) > max {
				return input[:max] + replacement
			}
			return input
		},
	}
	blocksTemplate, _ := template.New("blocks").Funcs(funcMap).Parse(`
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*<{{url . | escape}}|{{.Title | escape}}>*\n{{.Description | cutoff | escape}}"
			},
			"accessory": {
				"type": "image",
				"image_url": "{{.ImageURL | escape}}",
				"alt_text": "{{.Title | escape}}"
			}
		},
		{
			"type": "divider"
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": "*Rating*\n{{rating .}}"
				},
				{
					"type": "mrkdwn",
					"text": "*#Ratings*\n{{.RatingsCount}}"
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": "*Price*\n{{price .}}"
				}
			]
		},
		{
			"type": "divider"
		},
		{
			"type": "context",
			"elements": [
				{
					"type": "mrkdwn",
					"text": "*Something wrong with this result?*\nReact with :cry: to report and we'll look into it!"
				}
			]
		}
	]
	`)

	var buffer bytes.Buffer
	blocksTemplate.Execute(&buffer, p)

	fullJSONString := fmt.Sprintf(`
	{
		"channel": "%s",
		%s
	}`, channelId, buffer.String())

	return fullJSONString
}

func (s *Slack) OnMessage(cb OnMessageCallback) error {
	temp := s.typeToHandler[slackeventMessage]
	s.typeToHandler[slackeventMessage] = append(temp, func(emc *slackEventMessageContainer, w http.ResponseWriter, r *http.Request) {
		e := emc.Event
		for _, b := range e.Blocks {
			for _, parentElement := range b.Elements {
				for _, element := range parentElement.Elements {
					cb(s, &Message{
						ID:                   e.ClientMessageID,
						Content:              element.Url,
						MessageIsFromThisBot: false,
						Actions: &SlackMessageActions{
							event: &e,
							slack: s,
						},
					})
				}
			}
		}
	})
	return nil
}

func (s *Slack) OnProductProblemReport(OnProductProblemReportCallback) error {
	return nil
}

//Start an HTTP server and listen for Slack events
func (s *Slack) Start(port string) {
	http.ListenAndServe(port, s)
}

//ServeHTTP to implement http.Handler
func (eh *Slack) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message, error := ParseEventMessage(r.Body)

	if error != nil {
		fmt.Println(error)
		return
	}

	handlers := eh.typeToHandler[message.Event.Type]

	if handlers != nil {
		for _, h := range handlers {
			h(message, w, r)
		}
	}
}
