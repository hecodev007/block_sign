package models

type ImportKey struct {
	Address    string `json:"address"`
	AesPrivkey string `json:"aesPrivkey"`
	AesKey     string `json:"aesKey"`
}

type ImportKey2 struct {
	Privkey string `json:"privkey"`
	Address string `json:"address"`
}
