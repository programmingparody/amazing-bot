package chatapp

//OnMessageCallback should be called when a message is received on a session
type OnMessageCallback func(Session, *Message)

//OnProductProblemReportCallback should be called when a user reports a message. Typically through a reaction
type OnProductProblemReportCallback func(s Session, messageID string)

//Session handler of a chat
type Session interface {
	OnMessage(OnMessageCallback) error
	OnProductProblemReport(OnProductProblemReportCallback) error
}

//Actions to perform on a chat message
type Actions interface {
	Remove() error
	RespondWithProduct(*Product) (newMessageID string, e error)
}

//Message from a chat
type Message struct {
	ID                   string //Unique ID of the message
	Content              string
	MessageIsFromThisBot bool //Is this our own message (used for ignoring messages)
	Actions              Actions
}
