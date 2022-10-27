package crust

// import (
// 	"github.com/shopspring/decimal"

// 	"github.com/JFJun/bifrost-go/client"
// )

// type Client struct {
// 	*client.Client
// 	//C                  *gsrc.SubstrateAPI
// 	url string
// 	api string
// }

// func New(url string, api string) (c *Client, err error) {
// 	c = new(Client)
// 	c.url = url
// 	c.Client, err = client.New(url)
// 	// 初始化rpc客户端
// 	if err != nil {
// 		return nil, err
// 	}
// 	c.api = api
// 	c.SetPrefix(HydraDXPrefix)
// 	return c, nil
// }

// type AccountInfo struct {
// 	Nonce       decimal.Decimal `json:"nonce"`
// 	TokenSymbol string          `json:"tokenSymbol"`
// 	Free        decimal.Decimal `json:"free"`
// }

// //
// ///*
// //根据地址获取地址的账户信息，包括nonce以及余额等
// //*/
// //func (c *Client) GetAccountInfo(address string) (acc *AccountInfo, err error) {
// //	resp, err := http.Get(fmt.Sprintf("%s/accounts/%s/balance-info", c.api, address))
// //	if err != nil {
// //		return nil, err
// //	}
// //	defer resp.Body.Close()
// //	body, err := ioutil.ReadAll(resp.Body)
// //	if err != nil {
// //		return nil, err
// //	}
// //	acc = new(AccountInfo)
// //	err = json.Unmarshal(body, acc)
// //	if err != nil {
// //		return
// //	}
// //	return
// //}
