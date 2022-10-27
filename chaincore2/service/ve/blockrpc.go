package ve

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/common/log"
	"strings"
	//"github.com/vechain/thor/tx"
)

type Client struct {
	Nodeurl string
}

func NewRskBlock(node string) *Client {
	client := new(Client)
	client.Nodeurl = node
	return client
}

// 获取节点区块高度
func (c *Client) GetNodeHeight() (int64, error) {
	// 操作neo节点
	url := fmt.Sprintf("%s/blocks/best", c.Nodeurl)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}

	type heightS struct {
		Number int64 `json:"number"`
	}
	var h heightS
	err = json.Unmarshal(bytes, &h)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	return h.Number, nil
}

type BlockInfo struct {
	Height       int64    `json:"number"`
	Beneficiary  string   `json:"beneficiary"`
	GasLimit     int64    `json:"gasLimit"`
	GasUsed      int64    `json:"gasUsed"`
	ID           string   `json:"id"`
	IsTrunk      bool     `json:"isTrunk"`
	ParentID     string   `json:"parentID"`
	ReceiptsRoot string   `json:"receiptsRoot"`
	Signer       string   `json:"signer"`
	Size         int64    `json:"size"`
	StateRoot    string   `json:"stateRoot"`
	Timestamp    int64    `json:"timestamp"`
	TotalScore   int64    `json:"totalScore"`
	TxsFeatures  int64    `json:"txsFeatures"`
	TxsRoot      string   `json:"txsRoot"`
	Transactions []string `json:"transactions"`
}

//根据传入高度获取区块信息
func (c *Client) GetBlockInfoByHeight(height int64) (*BlockInfo, error) {
	var block BlockInfo

	// 操作节点
	url := fmt.Sprintf("%s/blocks/%d", c.Nodeurl, height)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	if err != nil {
		log.Error(err.Error())
		return &block, err
	}

	err = json.Unmarshal(bytes, &block)
	if err != nil {
		log.Error(err)
		return &block, err
	}
	return &block, nil
}

type TransactionResult struct {
	GasPayer string `json:"gasPayer"`
	GasUsed  int64  `json:"gasUsed"`
	Meta     struct {
		BlockID        string `json:"blockID"`
		BlockNumber    int64  `json:"blockNumber"`
		BlockTimestamp int64  `json:"blockTimestamp"`
		TxID           string `json:"txID"`
		TxOrigin       string `json:"txOrigin"`
	} `json:"meta"`
	Outputs []struct {
		ContractAddress interface{} `json:"contractAddress,omitempty"`
		Events          []struct {
			Address string   `json:"address"`
			Data    string   `json:"data"`
			Topics  []string `json:"topics"`
			//Topics  []thor.Bytes32 `json:"topics"`
		} `json:"events"`
		//Events tx.Events `json:"events"`
		Transfers []struct {
			Sender    string `json:"sender"`
			Recipient string `json:"recipient"`
			Amount    string `json:"amount"`
		} `json:"transfers"`
	} `json:"outputs"`
	Paid     string `json:"paid"`
	Reverted bool   `json:"reverted"`
	Reward   string `json:"reward"`
}

// 获取区块数据
func (c *Client) GetTransactionsByID(txID string) (*TransactionResult, error) {
	var result TransactionResult

	url := fmt.Sprintf("%s/transactions/%s/receipt", c.Nodeurl, txID)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	//beego.Debug(string(bytes))
	if err != nil {
		return &result, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		fmt.Println(err.Error())
		return &result, err
	}

	return &result, nil
}

type PendingStruct struct {
	BlockRef string `json:"blockRef"`
	ChainTag int64  `json:"chainTag"`
	Clauses  []struct {
		Data  string `json:"data"`
		To    string `json:"to"`
		Value string `json:"value"`
	} `json:"clauses"`
	Delegator    interface{} `json:"delegator"`
	DependsOn    interface{} `json:"dependsOn"`
	Expiration   int64       `json:"expiration"`
	Gas          int64       `json:"gas"`
	GasPriceCoef int64       `json:"gasPriceCoef"`
	ID           string      `json:"id"`
	Meta         struct {
		BlockID        string `json:"blockID"`
		BlockNumber    int64  `json:"blockNumber"`
		BlockTimestamp int64  `json:"blockTimestamp"`
	} `json:"meta"`
	Nonce  string `json:"nonce"`
	Origin string `json:"origin"`
	Size   int64  `json:"size"`
}

func (c *Client) GetTransactionsByPending(txID string, topics []string) (string, int64, int64, error) {
	var (
		result PendingStruct
	)

	url := fmt.Sprintf("%s/transactions/%s?pending=true", c.Nodeurl, txID)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	if err != nil {
		return "", 0, 0, err
	}
	//log.Debug(string(bytes))
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", 0, 0, err
	}

	for i := 0; i < len(result.Clauses); i++ {
		for _, v := range topics {
			hexaddr := v
			if strings.HasPrefix(v, "0x") {
				hexaddr = strings.TrimPrefix(v, "0x")
			}
			if strings.Contains(result.Clauses[i].Data, hexaddr) {
				return v, result.Gas, result.GasPriceCoef, nil
			}
		}
	}

	return "", result.Gas, result.GasPriceCoef, errors.New("has not get the addr")
}

// 获取节点区块高度
func getblock_count() (string, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "getblockcount",
		"params":  []interface{}{},
	}
	req.JSONBody(reqbody)
	respdata, err := req.String()
	return respdata, err
}

// 获取区块数据
func getblock_data(val interface{}) (string, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "getblock",
		"params":  []interface{}{val, 1},
	}
	req.JSONBody(reqbody)
	respdata, err := req.String()
	return respdata, err
}
