package dingding

type DdingStruct struct {
	ChatbotUserId  string      `json:"chatbotUserId"`  //机器人id
	ConversationId string      `json:"conversationId"` //
	SenderId       string      `json:"senderId"`       //人员ID
	SenderNick     string      `json:"senderNick"`     //昵称
	IsAdmin        bool        `json:"isAdmin"`        //是否是管理员
	CreateAt       int64       `json:"createAt"`       //创建时间
	Msgtype        string      `json:"msgtype"`        //text
	Text           MsgtypeText //text类型内容
}
type MsgtypeText struct {
	Content string `json:"content"`
}
