package dingbot

// MarkdownMessage is used to construct Markdown Message body
type MarkdownMessage struct {
	baseMessage
	Markdown Markdown `json:"markdown"`
	At       AtTag    `json:"at,omitempty"`
}

// Markdown contains basic information of a markdown message
type Markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// NewMarkdownMessage creates a markdown message
func NewMarkdownMessage(title string, text string, at AtTag) *MarkdownMessage {
	return &MarkdownMessage{
		baseMessage: baseMessage{"markdown"},
		Markdown: Markdown{
			Title: title,
			Text:  text,
		},
		At: at,
	}
}

// SimpleMarkdownMessage provides a simple Markdown message creation without tagging
func SimpleMarkdownMessage(text string) *MarkdownMessage {
	return NewMarkdownMessage("Robot Message", text, EmptyAtTag())
}

func (msg *MarkdownMessage) Send(accessToken string) error {
	defaultError := new(ResponseError)
	_, err := msg.baseMessage.GetClient(accessToken).New().BodyJSON(msg).ReceiveSuccess(defaultError)
	return err
}
