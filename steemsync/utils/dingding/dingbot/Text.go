package dingbot

// TextMessage is used to construct Text Message body
type TextMessage struct {
	baseMessage
	Text Text  `json:"text"`
	At   AtTag `json:"at,omitempty"`
}

// Text contains basic information of a Text message
type Text struct {
	Content string `json:"content"`
}

// NewTextMessage creates a text message
func NewTextMessage(content string, at AtTag) *TextMessage {
	return &TextMessage{
		baseMessage: baseMessage{MsgType: "text"},
		Text:        Text{Content: content},
		At:          at,
	}
}

// SimpleMarkdownMessage provides a quick way to create a text message
func SimpleTextMessage(content string) *TextMessage {
	return NewTextMessage(content, EmptyAtTag())
}

func (txt *TextMessage) Send(accessToken string) (*ResponseError, error) {
	defaultError := new(ResponseError)
	_, err := txt.baseMessage.GetClient(accessToken).New().BodyJSON(txt).ReceiveSuccess(defaultError)
	return defaultError, err
}
