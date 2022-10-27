package bo

type AddressInfo struct {
	Address          string `json:"address"`
	TransformAddress string `json:"transformAddress,omitempty"` //转换形态的address，比如bch对应的1地址
	SegWitAddr       string `json:"segWitAddr"`
	PrivateKey       string `json:"privateKey"`
}
