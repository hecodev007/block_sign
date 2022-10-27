package sol

import (
	"encoding/json"
	"strings"

	"github.com/portto/solana-go-sdk/client"
	"github.com/shopspring/decimal"
)

type Client struct {
	*RpcClient
	*client.Client
}

func NewClient(url string) *Client {
	cli := &Client{
		NewRpcClient(url, "", ""),
		client.NewClient(url),
	}
	return cli
}

type GetTokenAccountByOwner struct {
	Context struct {
		Slot int64 `json:"slot"`
	}
	Value []struct {
		Account struct {
			Data struct {
				Parsed struct {
					Info struct {
						IsNative    bool   `json:"isNative"`
						Mint        string `json:"mint"`
						Owner       string `json:"owner"`
						State       string `json:"state"`
						TokenAmount struct {
							Amount         decimal.Decimal `json:"amount"`
							Decimals       int64           `json:"decimals"`
							UiAmount       float64         `json:"uiAmount"`
							UiAmountString string          `json:"uiAmountString"`
						} `json:"tokenAmount"`
						Type string `json:"type"`
					} `json:"info"`
				} `json:"parsed"`
				Program string `json:"program"`
				Space   int64  `json:"space"`
			} `json:"data"`
			Executable bool   `json:"executable"`
			Lamports   int64  `json:"lamports"`
			Owner      string `json:"owner"`
			RentEpoch  int64  `json:"rentEpoch"`
		} `json:"account"`
		Pubkey string `json:"pubkey"`
	} `json:"value"`
}

func (cli *Client) BalanceOf(addr, contract string) (amount decimal.Decimal, tokenAddress string, decimals int64, err error) {
	params1 := make(map[string]string)
	params1["mint"] = contract
	params2 := make(map[string]string)
	params2["encoding"] = "jsonParsed"
	result := new(GetTokenAccountByOwner)
	err = cli.RpcClient.CallNoAuth("getTokenAccountsByOwner", result, addr, params1, params2)
	if err != nil {
		return
	}
	for _, v := range result.Value {
		if strings.ToLower(v.Account.Data.Parsed.Info.Mint) == strings.ToLower(contract) {
			return v.Account.Data.Parsed.Info.TokenAmount.Amount, v.Pubkey, v.Account.Data.Parsed.Info.TokenAmount.Decimals, nil
		}
	}
	return
}

type GetAccountInfo struct {
	Context struct {
		Slot int64 `json:"slot"`
	}
	Value struct {
		Data       json.RawMessage `json:"data"`
		Executable bool            `json:"executable"`
		Lamports   int64           `json:"lamports"`
		Owner      string          `json:"owner"`
		RentEpoch  int64           `json:"rentEpoch"`
	} `json:"value"`
}
type TokenAccountData struct {
	Parsed struct {
		Info struct {
			IsNative    bool   `json:"isNative"`
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         decimal.Decimal `json:"amount"`
				Decimals       int64           `json:"decimals"`
				UiAmount       float64         `json:"uiAmount"`
				UiAmountString string          `json:"uiAmountString"`
			} `json:"tokenAmount"`
			Type string `json:"type"`
		} `json:"info"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   int64  `json:"space"`
}

func (cli *Client) GetAccountInfo(addr string) (isTokenAddress bool, mint string, err error) {

	params2 := make(map[string]string)
	params2["encoding"] = "jsonParsed"
	result := new(GetAccountInfo)
	err = cli.RpcClient.CallNoAuth("getAccountInfo", result, addr, params2)
	if err != nil {
		return
	}
	if result.Value.Owner == "11111111111111111111111111111111" {
		return false, "", nil
	}
	resultData := new(TokenAccountData)
	if err = json.Unmarshal(result.Value.Data, resultData); err != nil {
		return
	}
	return true, resultData.Parsed.Info.Mint, nil
}
