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

type slackMessageActions struct {
	event *slackMessage
	slack *Slack
}

//Remove implementation for Actions
func (a *slackMessageActions) Remove() error {
	return nil
}

func (s *Slack) apiRequest(url string, jsonData []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.token))
	req.Header.Add("Content-type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	resData, error := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	return resData, error
}

//RespondWithProduct implementation for Actions
func (a *slackMessageActions) RespondWithProduct(p *Product) (string, error) {
	e := a.event
	s := a.slack

	data := slackRichTextJSONFromProduct(e.ChannelID, p, s.reportReactionCode)
	resData, _ := s.apiRequest("https://slack.com/api/chat.postMessage", []byte(data))

	var responseMessage slackEventMessageContainer
	json.Unmarshal(resData, &responseMessage)

	id := responseMessage.Message.TimeStamp
	channelID := responseMessage.Channel
	s.react(channelID, id, s.reportReactionCode)

	if len(s.myID) == 0 {
		s.myID = responseMessage.Message.UserID
	}
	return id, nil
}

type slackMessage struct {
	BotID           string `json:"bot_id"`
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
				URL  string `json:"url"`
			} `json:"Elements"`
		} `json:"Elements"`
	} `json:"blocks"`
	ChannelID string `json:"channel"`
	TimeStamp string `json:"ts"`
	Reaction  string `json:"reaction"`
	Item      struct {
		Type      string `json:"type"`
		Channel   string `json:"channel"`
		TimeStamp string `json:"ts"`
	}
}

type slackEventMessageContainer struct {
	Channel   string       `json:"channel"`
	TimeStamp string       `json:"ts"`
	Token     string       `json:"token"`
	Challenge string       `json:"challenge"`
	Event     slackMessage `json:"event"`
	Message   slackMessage `json:"message"`
}

func parseEventMessage(reader io.ReadCloser) (*slackEventMessageContainer, error) {
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
	slackeventMessage       = "message"
	slackeventReactionAdded = "reaction_added"
)

type slackEventHandlerFunc func(e *slackEventMessageContainer, w http.ResponseWriter, r *http.Request)

//Slack Session implementation
type Slack struct {
	typeToHandler      map[string][]slackEventHandlerFunc
	token              string
	reportReactionCode string
	myID               string
}

//NewSlackSession returns a Slack session that implements chatapp.Session
func NewSlackSession(token string, reportReactionCode string) *Slack {
	handlers := make(map[string][]slackEventHandlerFunc)
	handlers[slackeventMessage] = []slackEventHandlerFunc{}
	return &Slack{
		typeToHandler:      handlers,
		token:              token,
		reportReactionCode: reportReactionCode,
	}
}

func slackRichTextJSONFromProduct(channelID string, p *Product, negativeReaction string) string {
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
		"reportReaction": func() string {
			return negativeReaction
		},
	}
	//WARNING: Building JSON this way is at risk of failing randomly and cause validation errors at runtime
	//TODO: Create an anonymous struct to represent this instead, and use json.Marshal
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
					"text": "*Something wrong with this result?*\nReact with :{{reportReaction}}: to report and we'll look into it!"
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
	}`, channelID, buffer.String())

	return fullJSONString
}

func (s *Slack) react(channelID string, id string, reactionCode string) {
	data, _ := json.Marshal(struct {
		Channel      string `json:"channel"`
		ReactionCode string `json:"name"`
		ID           string `json:"timestamp"`
	}{
		Channel:      channelID,
		ReactionCode: reactionCode,
		ID:           id,
	})

	s.apiRequest("https://slack.com/api/reactions.add", data)
}

//OnMessage implements Session
func (s *Slack) OnMessage(cb OnMessageCallback) error {
	temp := s.typeToHandler[slackeventMessage]
	s.typeToHandler[slackeventMessage] = append(temp, func(emc *slackEventMessageContainer, w http.ResponseWriter, r *http.Request) {
		e := emc.Event

		for _, b := range e.Blocks {
			for _, parentElement := range b.Elements {
				for _, element := range parentElement.Elements {
					cb(s, &Message{
						ID:                   e.ClientMessageID,
						Content:              element.URL,
						MessageIsFromThisBot: len(emc.Message.BotID) != 0,
						Actions: &slackMessageActions{
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

//OnProductProblemReport implements Session
func (s *Slack) OnProductProblemReport(cb OnProductProblemReportCallback) error {
	temp := s.typeToHandler[slackeventMessage]
	s.typeToHandler[slackeventReactionAdded] = append(temp, func(emc *slackEventMessageContainer, w http.ResponseWriter, r *http.Request) {
		reaction := emc.Event.Reaction
		reactionFromBot := emc.Event.UserID == s.myID
		if reaction == s.reportReactionCode && !reactionFromBot {
			cb(s, emc.Event.Item.TimeStamp)
		}
	})
	return nil
}

//Start an HTTP server and listen for Slack events
func (s *Slack) Start(port string) {
	http.ListenAndServe(port, s)
}

//ServeHTTP to implement http.Handler
func (s *Slack) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message, error := parseEventMessage(r.Body)

	if error != nil {
		return
	}

	handlers := s.typeToHandler[message.Event.Type]

	if handlers != nil {
		for _, h := range handlers {
			h(message, w, r)
		}
	} else {
		w.Write([]byte(message.Challenge))
	}
}
