package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/codec"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
)

const avaxDecimal = 9

type AvaxUtxo struct {
	Address     string
	UtxoStr     string
	AmountFolat decimal.Decimal
}

func ParseUtxosBySortAsc(rawUtxo []string) (AvaxUnspentAsc, error) {
	avaxUtxos := make([]AvaxUtxo, 0)
	for _, v := range rawUtxo {
		avaxutxo, err := ParseUtxoToTpl(v)
		if err == nil {
			avaxUtxos = append(avaxUtxos, *avaxutxo)
		} else {
			log.Error(err.Error())
		}
	}
	log.Infof("avaxUtxos:%d", len(avaxUtxos))
	var sortUtxoTmp AvaxUnspentAsc
	sortUtxoTmp = append(sortUtxoTmp, avaxUtxos...)
	sort.Sort(sortUtxoTmp)
	log.Infof("sortUtxoTmp:%d", len(sortUtxoTmp))
	return sortUtxoTmp, nil
}

func ParseUtxosBySortDesc(rawUtxo []string) (AvaxUnspentDesc, error) {
	avaxUtxos := make([]AvaxUtxo, 0)
	for _, v := range rawUtxo {
		avaxutxo, err := ParseUtxoToTpl(v)
		if err == nil {
			avaxUtxos = append(avaxUtxos, *avaxutxo)
		} else {
			log.Error(err.Error())
		}
	}
	var sortUtxoTmp AvaxUnspentDesc
	sortUtxoTmp = append(sortUtxoTmp, avaxUtxos...)
	sort.Sort(sortUtxoTmp)
	return sortUtxoTmp, nil
}

func ShoToAddr(id ids.ShortID) (address string, err error) {
	//1主网地址验证
	address, err = formatting.FormatBech32(constants.NetworkIDToHRP[1], id.Bytes())
	address = "X-" + address
	return
}

func ParseUtxo(rawUtxo string) (*avax.UTXO, error) {
	fb := formatting.CB58{}
	fb.FromString(rawUtxo)
	c := codec.NewDefault()
	{
		c.RegisterType(&avm.BaseTx{})
		c.RegisterType(&avm.CreateAssetTx{})
		c.RegisterType(&avm.OperationTx{})
		c.RegisterType(&avm.ImportTx{})
		c.RegisterType(&avm.ExportTx{})
		c.RegisterType(&secp256k1fx.TransferInput{})
		c.RegisterType(&secp256k1fx.MintOutput{})
		c.RegisterType(&secp256k1fx.TransferOutput{})
		c.RegisterType(&secp256k1fx.MintOperation{})
		c.RegisterType(&secp256k1fx.Credential{})
	}
	utxo := &avax.UTXO{}
	if err := c.Unmarshal(fb.Bytes, utxo); err != nil {
		return nil, err
	}
	//utxojson, _ := json.Marshal(utxo)
	//fmt.Println(string(utxojson))
	return utxo, nil
}

func ParseUtxoToTpl(rawUtxo string) (*AvaxUtxo, error) {
	fb := formatting.CB58{}
	fb.FromString(rawUtxo)
	c := codec.NewDefault()
	{
		c.RegisterType(&avm.BaseTx{})
		c.RegisterType(&avm.CreateAssetTx{})
		c.RegisterType(&avm.OperationTx{})
		c.RegisterType(&avm.ImportTx{})
		c.RegisterType(&avm.ExportTx{})
		c.RegisterType(&secp256k1fx.TransferInput{})
		c.RegisterType(&secp256k1fx.MintOutput{})
		c.RegisterType(&secp256k1fx.TransferOutput{})
		c.RegisterType(&secp256k1fx.MintOperation{})
		c.RegisterType(&secp256k1fx.Credential{})
	}
	utxo := &avax.UTXO{}
	if err := c.Unmarshal(fb.Bytes, utxo); err != nil {
		return nil, err
	}
	output, ok := utxo.Out.(*secp256k1fx.TransferOutput)
	if !ok {
		return nil, errors.New("out format error")
	}
	if len(output.Addrs) == 0 {
		return nil, errors.New("addr format error")
	}
	addr, err := ShoToAddr(output.Addrs[0])
	if err != nil {
		return nil, fmt.Errorf("ShoToAddr error :%s", err.Error())
	}
	avaxUtxo := &AvaxUtxo{
		Address:     addr,
		UtxoStr:     rawUtxo,
		AmountFolat: decimal.NewFromInt(int64(output.Amount())).Shift(-1 * avaxDecimal),
	}
	return avaxUtxo, nil
}

func AvaxGetTxFee(host string) (int64, error) {
	host += "/ext/info"
	params := struct {
		Id      string `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Method  string `json:"method"`
	}{
		Id:      "test",
		Jsonrpc: "2.0",
		Method:  "info.getTxFee",
	}
	body, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("ReadAll err: %v", err)
	}
	ret := &struct {
		ID      string `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			TxFee string `json:"txFee"`
		} `json:"result"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	if err = json.Unmarshal(data, ret); err != nil {
		return 0, err
	}
	if ret.Error.Code != 0 {
		return 0, errors.New(ret.Error.Message)
	}
	return strconv.ParseInt(ret.Result.TxFee, 10, 64)
}
func AvaxGetUtxos(host string, address ...string) ([]string, error) {
	host += "/ext/bc/X"
	params := struct {
		Id      string `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  struct {
			Addresses []string `json:"addresses"`
			Limit     int      `json:"params"`
		} `json:"params"`
	}{
		Id:      "test",
		Jsonrpc: "2.0",
		Method:  "avm.getUTXOs",
		Params: struct {
			Addresses []string `json:"addresses"`
			Limit     int      `json:"params"`
		}{Addresses: address, Limit: 100},
	}
	body, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}
	ret := &struct {
		ID      string `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Utxos []string `json:"utxos"`
		} `json:"result"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	if err = json.Unmarshal(data, ret); err != nil {
		return nil, err
	}
	if ret.Error.Code != 0 {
		return nil, errors.New(ret.Error.Message)
	}
	return ret.Result.Utxos, nil
}

// unspents切片排序
type AvaxUnspentDesc []AvaxUtxo

//实现排序三个接口
//为集合内元素的总数
func (s AvaxUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s AvaxUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AvaxUnspentDesc) Less(i, j int) bool {
	return s[i].AmountFolat.GreaterThan(s[j].AmountFolat)
}

// unspents切片排序
type AvaxUnspentAsc []AvaxUtxo

//实现排序三个接口
//为集合内元素的总数
func (s AvaxUnspentAsc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s AvaxUnspentAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从小到大，最小金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s AvaxUnspentAsc) Less(i, j int) bool {
	return s[i].AmountFolat.LessThan(s[j].AmountFolat)
}
