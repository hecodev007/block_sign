package domain

type PhoneInfo struct {
	Tag       string `json:"tag,omitempty"`
	CodeName  string `json:"code_name,omitempty"`
	CodeValue string `json:"code_value,omitempty"`
}
