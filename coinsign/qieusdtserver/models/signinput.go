package models

type SignInput struct {
	Raw   string  `json:"raw_tx"`
	Txins []Txins `json:"txins"`
	Hash  string  `json:"hash,omitempty"`
	//PrivateKeys []string `json:"privateKeys,omitempty"`
	MchInfo
}
