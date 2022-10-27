package ada

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/onethefour/common/xutils"

	"github.com/shopspring/decimal"
)

type BlockResponse struct {
	Code              int    `json:"code"`
	Message           string `json:"message"`
	Block             Block  `json:"block"`
	OtherTransactions []struct {
		Hash string `json:"hash"`
	} `json:"other_transactions"`
}
type Block struct {
	BlockIdentifier struct {
		Index int64  `json:"index"`
		Hash  string `json:"hash"`
	} `json:"block_identifier"`
	Timestamp    int64          `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
}
type Transaction struct {
	TransactionIdentifier struct {
		Hash string `json:"hash"`
	} `json:"transaction_identifier"`
	Operations []Operation `json:"operations"`
}
type Operation struct {
	OperationIdentifier struct {
		Index int64 `json:"index"`
	} `json:"operation_identifier"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Account struct {
		Address string `json:"address"`
	}
	Amount struct {
		Value    decimal.Decimal `json:"value"`
		Currency struct {
			Symbol   string `json:"symbol"`
			Decimals int32  `json:"decimals"`
		} `json:"currency"`
	}
	CoinChange struct {
		CoinIdentifier struct {
			Identifier string `json:"identifier"`
		} `json:"coin_identifier"`
	} `json:"coin_change"`
	Metadata struct {
		TokenBundle []struct {
			PolicyId string `json:"policyId"`
			Tokens   []struct {
				Value    decimal.Decimal `json:"value"`
				Currency struct {
					Symbol  string `json:"symbol"`
					Decials int32  `json:"decials"`
				} `json:"currency"`
			} `json:"tokens"`
		} `json:"tokenBundle"`
	} `json:"metadata"`
}

func (rpc *RpcClient) GetBlockByHeight(height int64) (*Block, error) {
	return rpc.GetBlock(height)
}
func (rpc *RpcClient) GetBlock(height int64) (*Block, error) {
	params := fmt.Sprintf("{\"network_identifier\":{\"blockchain\":\"cardano\",\"network\":\"mainnet\"},\"block_identifier\":{\"index\":%v}}", height)
	body, err := rpc.post("/block", []byte(params))
	if err != nil {
		return nil, err
	}

	ret := new(BlockResponse)
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, err
	}
	//log.Info(len(ret.OtherTransactions))
	for k, _ := range ret.OtherTransactions {
		txret, err := rpc.GetRawTransaction(ret.OtherTransactions[k].Hash)
		if err != nil {
			return nil, err
		}
		//log.Info(ret.OtherTransactions[k].Hash)
		//if ret.OtherTransactions[k].Hash == "2da0ece648e8d147be35a267b53c6b41e61da037fb0ad6d97d3e0f9419a33efe" {
		//	log.Info(xutils.String(txret))
		//}
		if len(txret) == 1 {
			ret.Block.Transactions = append(ret.Block.Transactions, txret[0].Transaction)
		}
	}
	return &ret.Block, nil
}

type TransactionResponse struct {
	Block       []*TransactionRet `json:"transactions"`
	total_count int               `json:"total_count"`
}
type TransactionRet struct {
	BlockIdentifier struct {
		Index int64  `json:"index"`
		Hash  string `json:"hash"`
	} `json:"block_identifier"`
	Transaction *Transaction `json:"transaction"`
}

func (rpc *RpcClient) GetRawTransaction(txhash string) ([]*TransactionRet, error) {
	params := fmt.Sprintf("{\"network_identifier\":{\"blockchain\":\"cardano\",\"network\":\"mainnet\"},\"transaction_identifier\":{\"hash\":\"%v\"},\"success\":true}", txhash)
	body, err := rpc.post("/search/transactions", []byte(params))
	ret := new(TransactionResponse)
	//log.Info(string(body))
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, err
	}
	return ret.Block, nil
}

type CoinsResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Coins   []*Coin `json:"coins"`
}
type Coin struct {
	CoinIdentifier struct {
		Identifier string `json:"identifier"`
	} `json:"coin_identifier"`
	Amount struct {
		Value    decimal.Decimal `json:"value"`
		Currency struct {
			Symbol   string `json:"symbol"`
			Decimals int    `json:"decimals"`
		}
	}
}

func (rpc *RpcClient) Coins(addr string, mempool bool) ([]*Coin, error) {

	params := fmt.Sprintf("{\"network_identifier\":{\"blockchain\":\"cardano\",\"network\":\"mainnet\"},\"account_identifier\":{\"address\":\"%v\"},\"include_mempool\":%v}", addr, mempool)
	body, err := rpc.post("/account/coins", []byte(params))
	ret := new(CoinsResponse)
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, err
	}
	if ret.Code != 0 {
		return nil, errors.New(ret.Message)
	}
	return ret.Coins, nil

}

func (rpc *RpcClient) GetBlockCount() (int64, error) {
	ret, err := rpc.GetBalance("addr1v832ehavrtrr925kzuzlvkwmnyrk8ascz4qe22zef8lgskq4c93a9")
	if err != nil {
		return 0, err
	}
	return ret.BlockIdentifier.Index, nil
}

type BalanceResponse struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	BlockIdentifier struct {
		Index int64  `json:"index"`
		Hash  string `json:"hash"`
	} `json:"block_identifier"`
	Balances []*Balance `json:"balances"`
}

type Balance struct {
	Value    decimal.Decimal `json:"value"`
	Currency struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"currency"`
}

func (rpc *RpcClient) GetBalance(addr string) (*BalanceResponse, error) {
	params := fmt.Sprintf("{\"network_identifier\":{\"blockchain\":\"cardano\",\"network\":\"mainnet\"},\"account_identifier\":{\"address\":\"%v\"}}", addr)
	body, err := rpc.post("/account/balance", []byte(params))
	ret := new(BalanceResponse)
	if err = json.Unmarshal(body, ret); err != nil {
		return nil, err
	}
	if ret.Code != 0 {
		return nil, errors.New(ret.Message)
	}
	return ret, nil
}

func (rpc *RpcClient) BalanceOf(addr string, token string, decimals int) (decimal.Decimal, error) {
	balances, err := rpc.GetBalance(addr)
	if err != nil {
		return decimal.Decimal{}, err
	}
	if token == "" {
		token = "ADA"
	}
	println(xutils.String(balances))
	var amount decimal.Decimal
	for _, balance := range balances.Balances {
		if strings.ToUpper(balance.Currency.Symbol) == strings.ToUpper(token) {
			if decimals == 0 {
				decimals = balance.Currency.Decimals
			}
			if decimals != balance.Currency.Decimals {
				return amount, fmt.Errorf("严重错误,%v 精度不一致 数据库精度:%v 链上精度:%v", token, decimals, balance.Currency.Decimals)
			}
			amount = amount.Add(balance.Value)
		}
	}
	return amount, nil
}
