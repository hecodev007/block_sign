package bo

type AddressInfo struct {
	Address    string `json:"address"`
	SegWitAddr string `json:"segWitAddr"`
	PrivateKey string `json:"privateKey"`
	PrivateHex string `json:"privateHex"`
}
