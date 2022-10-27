package dingbot

// FeedCardMessage is used to construct FeedCard Message body
type FeedCardMessage struct {
	*baseMessage
	FeedCard *FeedCard `json:"feedCard"`
}

type FeedCard struct {
	Links []*FeedCardLink `json:"links"`
}
type FeedCardLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

type FeedCardBuilder struct {
	Links []*FeedCardLink
}

func NewFeedCardBuilder() *FeedCardBuilder {
	return &FeedCardBuilder{}
}

func (builder *FeedCardBuilder) addLink(title string, msgURL string, picURL string) *FeedCardBuilder {
	if builder.Links != nil || len(builder.Links) <= 0 {
		builder.Links = []*FeedCardLink{{
			Title:      title,
			MessageURL: msgURL,
			PicURL:     picURL,
		}}
	} else {
		builder.Links = append(
			builder.Links,
			&FeedCardLink{
				Title:      title,
				MessageURL: msgURL,
				PicURL:     picURL,
			},
		)
	}

	return builder
}

func (builder *FeedCardBuilder) build() *FeedCardMessage {
	return &FeedCardMessage{
		baseMessage: &baseMessage{MsgType: "feedCard"},
		FeedCard:    &FeedCard{Links: builder.Links},
	}
}

func (msg *FeedCardMessage) Send(accessToken string) error {
	defaultError := new(ResponseError)
	_, err := msg.baseMessage.GetClient(accessToken).New().BodyJSON(msg).ReceiveSuccess(defaultError)
	return err
}
