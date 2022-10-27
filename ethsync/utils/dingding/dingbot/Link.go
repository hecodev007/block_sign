package dingbot

// LinkMessage is used to construct Link Message body
type LinkMessage struct {
	*baseMessage
	Link *Link `json:"link"`
}

// Link contains basic information for a Link message
type Link struct {
	Text       string `json:"text"`
	Title      string `json:"title"`
	PicURL     string `json:"picUrl"`
	MessageURL string `json:"messageUrl"`
}

// NewLinkMessage help create a Link Message
func NewLinkMessage(text string, title string, picURL string, msgURL string) *LinkMessage {
	return &LinkMessage{
		baseMessage: &baseMessage{MsgType: "link"},
		Link: &Link{
			Text:       text,
			Title:      title,
			PicURL:     picURL,
			MessageURL: msgURL,
		},
	}
}

func (msg *LinkMessage) Send(accessToken string) error {
	defaultError := new(ResponseError)
	_, err := msg.baseMessage.GetClient(accessToken).New().BodyJSON(msg).ReceiveSuccess(defaultError)
	return err
}
