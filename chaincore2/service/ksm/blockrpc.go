package ksm

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
)

type KsmBlockStruct struct {
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Block struct {
			Extrinsics []string `json:"extrinsics"`
			Header     struct {
				Digest struct {
					Logs []string `json:"logs"`
				} `json:"digest"`
				ExtrinsicsRoot string `json:"extrinsicsRoot"`
				Number         string `json:"number"`
				ParentHash     string `json:"parentHash"`
				StateRoot      string `json:"stateRoot"`
			} `json:"header"`
		} `json:"block"`
		Justification interface{} `json:"justification"`
	} `json:"result"`
}

type KsmBlock struct {
	Nodeurl string
}

func NewKsmBlock(node string) *KsmBlock {
	ksm := new(KsmBlock)
	ksm.Nodeurl = node
	return ksm
}

func (ksm *KsmBlock) GetblockCount() (int64, error) {
	req := httplib.Post(ksm.Nodeurl) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "chain_getBlock",
		"params":  []interface{}{},
	}
	req.JSONBody(reqbody)
	//respdata, err := req.String()
	bytes, err := req.Bytes()
	if err != nil {
		return 0, err
	}
	var ksmB KsmBlockStruct
	err = json.Unmarshal(bytes, &ksmB)
	if err != nil {
		return 0, err
	}

	currentheight := common.StrBaseToInt(ksmB.Result.Block.Header.Number, 16)
	return int64(currentheight), nil
}

func (ksm *KsmBlock) GetBlockData(hash string) (KsmBlockStruct, error) {
	var ksmB KsmBlockStruct

	req := httplib.Post(ksm.Nodeurl)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "chain_getBlock",
		"params":  []interface{}{hash},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return ksmB, err
	}

	err = json.Unmarshal(bytes, &ksmB)
	if err != nil {
		return ksmB, err
	}

	return ksmB, nil
}

type HashResult struct {
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// 获取区块数据
func (ksm *KsmBlock) GethashkByHeight(height int64) (string, error) {
	// 操作neo节点
	req := httplib.Post(ksm.Nodeurl) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "chain_getBlockHash",
		"params":  []interface{}{height},
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if err != nil {
		return "", err
	}
	var hash HashResult
	err = json.Unmarshal(bytes, &hash)
	if err != nil {
		return "", err
	}
	return hash.Result, err
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

//type BlockTransMethods struct {
//	Extrinsics []struct {
//		Args   interface{} `json:"args"`
//		Method string      `json:"method"`
//	} `json:"extrinsics"`
//	Hash       string `json:"hash"`
//	ParentHash string `json:"parentHash"`
//}

//func GetBlockTransMethodsByHeight(height int64) (*BlockTransMethods, error) {
//	// 操作neo节点
//	url := fmt.Sprintf("%s/blocks/%d", beego.AppConfig.String("nodeurl2"), height)
//	req := httplib.Get(url)
//	bytes, err := req.Bytes()
//	if err != nil {
//		log.Error(err.Error())
//		return nil, err
//	}
//	log.Debug(string(bytes))
//	var trans BlockTransMethods
//	err = json.Unmarshal(bytes, &trans)
//	if err != nil {
//		log.Error(err)
//		return &trans, err
//	}
//	return &trans, nil
//}

func GetBlockTransMethodsByHeight(height int64) (*BlockTransStruct, error) {
	// 操作neo节点
	url := fmt.Sprintf("%s/blocks/%d", beego.AppConfig.String("nodeurl2"), height)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	//log.Debug(string(bytes))
	var trans BlockTransStruct
	err = json.Unmarshal(bytes, &trans)
	if err != nil {
		log.Error(err)
		log.Error(string(bytes))
		return &trans, err
	}
	return &trans, nil
}

//type BlockTransStruct struct {
//	Extrinsics []struct {
//		Args   interface{} `json:"args"`
//		Events []struct {
//			Data   []interface{} `json:"data"`
//			Method string        `json:"method"`
//		} `json:"events"`
//		Hash string `json:"hash"`
//		Info struct {
//			Class      string `json:"class"`
//			PartialFee string `json:"partialFee"`
//			Weight     string `json:"weight"`
//		} `json:"info"`
//		Method    struct{
//			Pallet string `json:"pallet"`
//			Method string `json:"method"`
//		} `json:"method"`
//		Nonce     string `json:"nonce"`
//		PaysFee   bool   `json:"paysFee"`
//		Signature struct {
//			Signature string `json:"signature"`
//			Signer    string `json:"signer"`
//		} `json:"signature"`
//		Success interface{} `json:"success"`
//		Tip     string      `json:"tip"`
//	} `json:"extrinsics"`
//	Hash       string `json:"hash"`
//	ParentHash string `json:"parentHash"`
//}
//============struct  start=================

type BlockTransStruct struct {
	Number         string       `json:"number"`
	Hash           string       `json:"hash"`
	ParentHash     string       `json:"parentHash"`
	StateRoot      string       `json:"stateRoot"`
	ExtrinsicsRoot string       `json:"extrinsicsRoot"`
	AuthorID       string       `json:"authorId"`
	Logs           []Logs       `json:"logs"`
	OnInitialize   OnInitialize `json:"onInitialize"`
	Extrinsics     []Extrinsics `json:"extrinsics"`
	OnFinalize     OnFinalize   `json:"onFinalize"`
	Finalized      bool         `json:"finalized"`
}
type Logs struct {
	Type  string   `json:"type"`
	Index string   `json:"index"`
	Value []string `json:"value"`
}
type OnInitialize struct {
	Events []interface{} `json:"events"`
}
type Method struct {
	Pallet string `json:"pallet"`
	Method string `json:"method"`
}

type Data struct {
	Weight  string `json:"weight"`
	Class   string `json:"class"`
	PaysFee string `json:"paysFee"`
}
type Events struct {
	Method Method `json:"method"`
	//Data []Data `json:"data"`
}
type Dest struct {
	ID string `json:"Id"`
}
type Args struct {
	Dest  interface{} `json:"dest"`
	Value string      `json:"value"`
	Now   string      `json:"now"`
}
type Info struct {
	Weight     string `json:"weight"`
	Class      string `json:"class"`
	PartialFee string `json:"partialFee"`
}
type Extrinsics struct {
	Method    Method      `json:"method"`
	Signature Signature   `json:"signature"`
	Nonce     interface{} `json:"nonce"`
	Args      Args        `json:"args,omitempty"`
	Tip       interface{} `json:"tip"`
	Hash      string      `json:"hash"`
	Info      Info        `json:"info,omitempty"`
	Events    []Events    `json:"events"`
	Success   bool        `json:"success"`
	PaysFee   bool        `json:"paysFee"`
}
type OnFinalize struct {
	Events []interface{} `json:"events"`
}

type Signature struct {
	Signature string `json:"signature"`
	Signer    struct {
		Id string `json:"Id"`
	} `json:"signer"`
}

//============struct  end=================

func GetBlockTransByHeight(height int64) (*BlockTransStruct, error) {
	url := fmt.Sprintf("%s/blocks/%d", beego.AppConfig.String("nodeurl2"), height)
	req := httplib.Get(url)
	bytes, err := req.Bytes()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//log.Debug(string(bytes))

	var trans BlockTransStruct
	err = json.Unmarshal(bytes, &trans)
	if err != nil {
		log.Error(err)
		return &trans, err
	}
	return &trans, nil
}
