package chatapp

type OnMessageCallback func(Session, *Message)
type OnProductProblemReportCallback func(s Session, messageID string)

type Session interface {
	OnMessage(OnMessageCallback) error
	OnProductProblemReport(OnProductProblemReportCallback) error
}

type Actions interface {
	Remove() error
	RespondWithProduct(*Product) (newMessageID string, e error)
}

type Message struct {
	ID                   string
	Content              string
	MessageIsFromThisBot bool
	Actions              Actions
}
