package dingbot

// SingleActionCardMessage is used to construct ActionCard Message body
type ActionCardMessage struct {
	baseMessage
	SingleActionCard SingleActionCard `json:"actionCard,omitempty"`
	MultiActionCard  MultiActionCard  `json:"actionCard,omitempty"`
}

type baseActionCard struct {
	Title          string `json:"title"`
	Text           string `json:"text"`
	HideAvatar     string `json:"hideAvatar,omitempty"`
	BtnOrientation string `json:"btnOrientation,omitempty"`
}

// SingleActionCard is an ActionCard with single button
type SingleActionCard struct {
	baseActionCard
	SingleTitle string `json:"singleTitle,omitempty"`
	SingleURL   string `json:"singleURL,omitempty"`
}

// Button contains information of action button
type Button struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// MultiActionCard is an ActionCard with multiple buttons
type MultiActionCard struct {
	baseActionCard
	Buttons []Button `json:"btns,omitempty"`
}

// NewSingleActionCard creates an ActionCard Message with a single button
func NewSingleActionCard(
	title string,
	text string,
	hideAvatar string,
	buttonOrientation string,
	buttonTitle string,
	buttonURL string,
) *ActionCardMessage {
	return &ActionCardMessage{
		baseMessage: baseMessage{MsgType: "actionCard"},
		SingleActionCard: SingleActionCard{
			baseActionCard: baseActionCard{
				Title:          title,
				Text:           text,
				HideAvatar:     hideAvatar,
				BtnOrientation: buttonOrientation,
			},
			SingleTitle: buttonTitle,
			SingleURL:   buttonURL,
		},
	}
}

// NewMultiActionCardBuilder creates an ActionCard Message with multiple buttons
func NewMultiActionCardBuilder(
	title string,
	text string,
	hideAvatar string,
	buttonOrientation string,
	buttons []Button,
) *ActionCardMessage {
	return &ActionCardMessage{
		baseMessage: baseMessage{MsgType: "actionCard"},
		MultiActionCard: MultiActionCard{
			baseActionCard: baseActionCard{
				Title:          title,
				Text:           text,
				HideAvatar:     hideAvatar,
				BtnOrientation: buttonOrientation,
			},
			Buttons: buttons,
		},
	}
}

func (msg *ActionCardMessage) Send(accessToken string) error {
	defaultError := new(ResponseError)
	_, err := msg.baseMessage.GetClient(accessToken).New().BodyJSON(msg).ReceiveSuccess(defaultError)
	return err
}
