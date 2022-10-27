package model

//托管后台


type BindMchRequest struct {
	Address       string `json:"address"`
	ApiKey      string `json:"api_key"`
	WhiteIp  string `json:"white_ip"`
	CoinName      string `json:"coin_name"`
}


type RegisterMchRequest struct {
	ApiKey       string `json:"api_key"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	CompanyImg string `json:"company_img"`
}

type GenerateTransportKeyType string

var (
	TypeGtkDeploy       GenerateTransportKeyType = "DEPLOY"
	TypeGtkGenerateAddr GenerateTransportKeyType = "GENERATE_ADDR"
)
type GenerateTransportKeyHandleRequest struct {
	CoinCode       string                   `json:"coinCode"`
	Mch            string                   `json:"mch"`
	Count          int64                    `json:"count"`
	//Type           GenerateTransportKeyType `json:"type"`
	//AssignSignerNo string                   `json:"assignSignerNo"`
	//SignerList     []string                 `json:"signerList"`
}


type BaseResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type AddressRsp struct {
	Mch    string         `json:"mch"`
	BatchNo string      `json:"batchNo"`
	CoinCode    string `json:"coinCode"`
	Address  []string `json:"address"`
}

