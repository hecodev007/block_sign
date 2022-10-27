package ada

import (
	"adasign/common/conf"
	"adasign/common/log"
	"adasign/common/validator"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/onethefour/common/xutils"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/coinbase/rosetta-sdk-go/client"
)

var MinAdaMap = map[string]uint64{"8a1cfae21368b8bebbbed9800fec304e95cce39a2a57dc35e2e3ebaa-4d494c4b-0": 1344798}
var DefaultMinAda = uint64(1400000)
var NetworkIdentifier = &types.NetworkIdentifier{
	Blockchain: "cardano",
	Network:    "mainnet",
}

func Sign(params *validator.SignParams) (txhash string, rawtx string, err error) {
	ctx := context.Background()
	clientCfg := client.NewConfiguration(
		conf.GetConfig().Node.Url,
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	//
	seeds := make(map[string]string)
	var totalInput = make(map[string]uint64)
	var totalOut = make(map[string]uint64)
	var inputops []*types.Operation
	var outputops []*types.Operation
	var changeops []*types.Operation
	for k, _ := range params.Ins {
		tmpop, err := NewInputOperation(cli, params.Ins[k].FromTxid, int(params.Ins[k].FromIndex), totalInput)
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		seeds[params.Ins[k].FromAddr] = params.Ins[k].FromPrivkey
		inputops = append(inputops, tmpop)
	}
	log.Info(xutils.String(totalInput))
	log.Info(xutils.String(inputops))
	//address,token,amount
	var outputParams = make(map[string]map[string]uint64)
	for k, _ := range params.Outs {
		if outputParams[params.Outs[k].ToAddr] == nil {
			outputParams[params.Outs[k].ToAddr] = make(map[string]uint64)
		}
		if params.Outs[k].Token == "" || params.Outs[k].Token == "ADA" {
			outputParams[params.Outs[k].ToAddr]["ADA"] += uint64(params.Outs[k].ToAmountInt64)
			totalOut["ADA"] += uint64(params.Outs[k].ToAmountInt64)
			if params.Outs[k].ToAmountInt64 != 0 && params.Outs[k].ToAmountInt64 < 1000000 {
				return "", "", errors.New("发送的ada 额度不能小于 1.000000")
			}
		} else {
			outputParams[params.Outs[k].ToAddr][params.Outs[k].Token] += uint64(params.Outs[k].ToAmountInt64)
			totalOut[params.Outs[k].Token] += uint64(params.Outs[k].ToAmountInt64)

			minAda, ok := MinAdaMap[params.Outs[k].Token]
			if !ok {
				minAda = DefaultMinAda
			}
			outputParams[params.Outs[k].ToAddr]["ADA"] += minAda
			totalOut["ADA"] += minAda
		}
	}

	for k, _ := range totalOut {
		if totalInput[k] < totalOut[k] {
			return "", "", errors.New(k + " 输入额度小于输出额度")
		}
		totalInput[k] -= totalOut[k]
		if totalInput[k] == 0 {
			delete(totalInput, k)
		}
	}
	for k, _ := range outputParams {
		tmpop, err := NewOutputsOperation(k, outputParams[k])
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		outputops = append(outputops, tmpop)
	}

	if changeops, err = ToChangeOutputsOperation(params.Change, totalInput); err != nil {
		return "", "", err
	}

	var ops []*types.Operation
	ops = append(ops, inputops...)
	ops = append(ops, outputops...)
	ops = append(ops, changeops...)

	preprocessRequest := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: NetworkIdentifier,
		Operations:        ops,
	}
	log.Info(xutils.String(ops))
	preresp, cerr, err := cli.ConstructionAPI.ConstructionPreprocess(ctx, preprocessRequest)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}
	var publickeys []*types.PublicKey
	for k, _ := range seeds {
		//		log.Info(seeds[k])
		tmpkp, err := ToKeyPire(seeds[k])
		if err != nil {
			log.Info(err.Error())
			return "", "", err
		}
		publickeys = append(publickeys, tmpkp.PublicKey)
	}
	req := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "cardano",
			Network:    "mainnet",
		},
		Options:    preresp.Options,
		PublicKeys: publickeys,
	}
	MetaDataresp, cerr, err := cli.ConstructionAPI.ConstructionMetadata(context.Background(), req)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}
	if len(MetaDataresp.SuggestedFee) != 1 {
		return "", "", errors.New("ConstructionMetadata 返回异常,签名服务需要升级")
	}
	log.Info(xutils.String(MetaDataresp))
	fee, err := strconv.Atoi(MetaDataresp.SuggestedFee[0].Value)
	log.Info("fee:", fee)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	}
	if fee < 171221 {
		fee = 171221
	}
	fee += 100 * (len(ops))
	if totalInput["ADA"] == 0 || totalInput["ADA"] < uint64(fee) {
		return "", "", fmt.Errorf("手续费不够%v > %v", fee, totalInput["ADA"])
	}
	log.Info("真实 fee:", fee, totalInput["ADA"])
	totalInput["ADA"] -= uint64(fee)
	if totalInput["ADA"] < 1000000 { //低于1个ada发不出去
		delete(totalInput, "ADA")
	}
	log.Info("真实 fee:", fee, totalInput["ADA"])
	ops = ops[0:0]
	ops = append(ops, inputops...)
	ops = append(ops, outputops...)
	log.Info("change output ", xutils.String(totalInput))
	if len(totalInput) > 0 {
		if changeops, err = ToChangeOutputsOperation(params.Change, totalInput); err != nil {
			return "", "", err
		}
		ops = append(ops, changeops...)
	}

	preresp, cerr, err = cli.ConstructionAPI.ConstructionPreprocess(ctx, preprocessRequest)
	if err != nil {
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}
	MetaDataresp, cerr, err = cli.ConstructionAPI.ConstructionMetadata(context.Background(), req)
	if err != nil {
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}
	md := MetaDataresp.Metadata

	if _,ok := md["ttl"]; ok {
		t :=  md["ttl"]
		ttl := GetIntFromInterface(t)
		ttlStr := fmt.Sprintf("%v", ttl + 5000)
		MetaDataresp.Metadata["ttl"] =ttlStr
	}

	payloadRrqust := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: NetworkIdentifier,
		Operations:        ops,
		Metadata:          MetaDataresp.Metadata,
	}
	payloadresp, cerr, err := cli.ConstructionAPI.ConstructionPayloads(ctx, payloadRrqust)
	if err != nil {
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	} else if len(payloadresp.Payloads) <= 0 {
		//log.Info(xutils.String(payloadresp))
		return "", "", errors.New("ConstructionPayloads 返回异常,签名需要升级")
	}

	var Signatures []*types.Signature

	for k, _ := range payloadresp.Payloads {
		seed, ok := seeds[payloadresp.Payloads[k].AccountIdentifier.Address]
		if !ok {
			return "", "", errors.New(payloadresp.Payloads[k].AccountIdentifier.Address + " 地址不存在")
		}
		tmpkp, _ := ToKeyPire(seed)
		signer, _ := tmpkp.Signer()
		sigerbytes, err := signer.Sign(payloadresp.Payloads[k], "ed25519")
		if err != nil {
			return "", "", err
		}
		signature := &types.Signature{
			SigningPayload: payloadresp.Payloads[k],
			PublicKey:      tmpkp.PublicKey,
			SignatureType:  "ed25519",
			Bytes:          sigerbytes.Bytes,
		}
		Signatures = append(Signatures, signature)
	}
	//
	//for k, _ := range seeds {
	//	tmpkp, _ := ToKeyPire(seeds[k])
	//	signer, _ := tmpkp.Signer()
	//	sigerbytes, err := signer.Sign(payloadresp.Payloads[0], "ed25519")
	//	if err != nil {
	//		return "", "", err
	//	}
	//	signature := &types.Signature{
	//		SigningPayload: payloadresp.Payloads[0],
	//		PublicKey:      tmpkp.PublicKey,
	//		SignatureType:  "ed25519",
	//		Bytes:          sigerbytes.Bytes,
	//	}
	//	Signatures = append(Signatures, signature)
	//}

	combileRequest := &types.ConstructionCombineRequest{
		NetworkIdentifier:   NetworkIdentifier,
		UnsignedTransaction: payloadresp.UnsignedTransaction,
		Signatures:          Signatures,
	}

	combileResp, cerr, err := cli.ConstructionAPI.ConstructionCombine(ctx, combileRequest)
	if err != nil {
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}
	hashRequest := &types.ConstructionHashRequest{
		NetworkIdentifier: NetworkIdentifier,
		SignedTransaction: combileResp.SignedTransaction,
	}
	hashResp, cerr, err := cli.ConstructionAPI.ConstructionHash(ctx, hashRequest)
	if err != nil {
		return "", "", err
	} else if cerr != nil {
		return "", "", errors.New(cerr.Message)
	}

	return hashResp.TransactionIdentifier.Hash, combileResp.SignedTransaction, nil
}


func GetIntFromInterface(in interface{}) (out int64) {
	switch in.(type) {
	case float64:
		f := in.(float64)
		out = int64(f)
	case string:
		f := in.(string)
		out, _ = strconv.ParseInt(f, 10, 64)
	case int64:
		out = in.(int64)
	case int:
		f := in.(int)
		out = int64(f)
	}
	return
}

type TokenMeta struct {
	TokenBundle []TokenBundle `json:"tokenBundle"`
}
type TokenBundle struct {
	PolicyId string  `json:"policyId"`
	Tokens   []Token `json:"tokens"`
}
type Token struct {
	Value    string   `json:"value"`
	Currency Currency `json:"currency"`
}
type Currency struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

func NewInputOperation(cli *client.APIClient, txid string, index int, totalInput map[string]uint64) (*types.Operation, error) {
	params := &types.SearchTransactionsRequest{
		NetworkIdentifier: NetworkIdentifier,
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txid,
		}}
	resp, cerr, err := cli.SearchAPI.SearchTransactions(context.Background(), params)
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return nil, errors.New(cerr.Message)
	}
	if len(resp.Transactions) != 1 {
		return nil, errors.New("input交易" + txid + " 没找到")
	}
	if len(resp.Transactions[0].Transaction.Operations) <= index {
		return nil, errors.New("input交易" + txid + " output没找到")
	}
	var oper *types.Operation
	for k, _ := range resp.Transactions[0].Transaction.Operations {
		if resp.Transactions[0].Transaction.Operations[k].CoinChange != nil &&
			resp.Transactions[0].Transaction.Operations[k].CoinChange.CoinIdentifier.Identifier == fmt.Sprintf("%v:%v", txid, index) {
			oper = resp.Transactions[0].Transaction.Operations[k]
			break
		}
	}

	//oper := resp.Transactions[0].Transaction.Operations[index]
	ret := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index:        0,
			NetworkIndex: nil,
		},
		Type: "input",
		Account: &types.AccountIdentifier{
			Address: oper.Account.Address,
		},
		Amount: oper.Amount,
		CoinChange: &types.CoinChange{
			CoinIdentifier: &types.CoinIdentifier{
				Identifier: fmt.Sprintf("%v:%v", txid, index),
			},
			CoinAction: "coin_spent",
		},
		Metadata: nil,
	}
	valueInt, err := strconv.Atoi(ret.Amount.Value)
	if err != nil {
		return nil, err
	}
	totalInput["ADA"] = totalInput["ADA"] + uint64(valueInt)

	if ret.Amount.Value != "" && ret.Amount.Value != "0" {
		ret.Amount.Value = "-" + ret.Amount.Value
	}

	if oper.Metadata != nil && len(oper.Metadata) != 0 {
		metaData, err := json.Marshal(oper.Metadata)
		if err != nil {
			return nil, err
		}
		metaInst := new(TokenMeta)
		err = json.Unmarshal(metaData, metaInst)
		if err != nil {
			return nil, err
		}
		for k, _ := range metaInst.TokenBundle {
			//if len(metaInst.TokenBundle) != 1 {
			//	return nil, errors.New("input.TokenBundle 解析错误,签名服务需要升级")
			//}
			if len(metaInst.TokenBundle[k].Tokens) != 1 {
				return nil, errors.New("input.TokenBundle.Tokens 解析错误,签名服务需要升级")
			}
			tokenAssert := fmt.Sprintf("%v-%v-%v", metaInst.TokenBundle[k].PolicyId, metaInst.TokenBundle[k].Tokens[0].Currency.Symbol, metaInst.TokenBundle[k].Tokens[0].Currency.Decimals)
			tokenValue, err := strconv.Atoi(metaInst.TokenBundle[k].Tokens[0].Value)
			if err != nil {
				return nil, err
			}
			if tokenValue < 0 {
				return nil, errors.New("input.TokenBundle.Tokens.value 解析错误,签名服务需要升级")
			}
			totalInput[tokenAssert] += uint64(tokenValue)

			metaInst.TokenBundle[k].Tokens[0].Value = "-" + metaInst.TokenBundle[k].Tokens[0].Value
		}
		metaData, _ = json.Marshal(metaInst)

		ret.Metadata = make(map[string]interface{})
		json.Unmarshal(metaData, &ret.Metadata)
	}
	return ret, nil
}

//token := policyId-symbol-decimals
func NewOutputOperation(toaddr string, adamount uint64, token string, tokenAmount uint64) (*types.Operation, error) {
	ret := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index:        0,
			NetworkIndex: nil,
		},
		Type: "output",
		Account: &types.AccountIdentifier{
			Address:    toaddr,
			SubAccount: nil,
			Metadata:   nil,
		},
		Amount:   nil,
		Metadata: make(map[string]interface{}),
	}
	if adamount != 0 {
		ret.Amount = &types.Amount{
			Value: fmt.Sprintf("%v", adamount),
			Currency: &types.Currency{
				Symbol:   "ADA",
				Decimals: 6,
				Metadata: nil,
			},
			Metadata: nil,
		}
	}
	if token != "" && tokenAmount != 0 {
		tokeninfos := strings.Split(token, "-")
		if len(tokeninfos) != 3 {
			return nil, errors.New("token 地址格式错误:" + token)
		}
		policyId := tokeninfos[0]
		symbol := tokeninfos[1]
		decimals, err := strconv.Atoi(tokeninfos[2])
		if err != nil {
			return nil, err
		}
		metaStr := fmt.Sprintf("{\"tokenBundle\":[{\"policyId\":\"%v\",\"tokens\":[{\"value\":\"%v\",\"currency\":{\"symbol\":\"%v\",\"decimals\":%v}}]}]}", policyId, tokenAmount, symbol, decimals)
		ret.Metadata = make(map[string]interface{})
		err = json.Unmarshal([]byte(metaStr), &ret.Metadata)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

//token := policyId-symbol-decimals
func NewOutputsOperation(toaddr string, tokens map[string]uint64) (*types.Operation, error) {
	ret := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index:        0,
			NetworkIndex: nil,
		},
		Type: "output",
		Account: &types.AccountIdentifier{
			Address:    toaddr,
			SubAccount: nil,
			Metadata:   nil,
		},
		Amount: &types.Amount{
			Value: "0",
			Currency: &types.Currency{
				Symbol:   "ADA",
				Decimals: 6,
				Metadata: nil,
			},
			Metadata: nil,
		},
		Metadata: make(map[string]interface{}),
	}
	adamount := tokens["ADA"]
	if adamount != 0 {
		ret.Amount = &types.Amount{
			Value: fmt.Sprintf("%v", adamount),
			Currency: &types.Currency{
				Symbol:   "ADA",
				Decimals: 6,
				Metadata: nil,
			},
			Metadata: nil,
		}
	}
	metaDataInst := new(TokenMeta)
	for token, tokenAmount := range tokens {
		if token == "ADA" {
			continue
		}
		if token != "" && tokenAmount != 0 {
			tokeninfos := strings.Split(token, "-")
			if len(tokeninfos) != 3 {
				return nil, errors.New("token 地址格式错误:" + token)
			}
			policyId := tokeninfos[0]
			symbol := tokeninfos[1]
			decimals, err := strconv.Atoi(tokeninfos[2])
			if err != nil {
				return nil, err
			}
			tokenBundleInst := TokenBundle{}
			metaStr := fmt.Sprintf("{\"policyId\":\"%v\",\"tokens\":[{\"value\":\"%v\",\"currency\":{\"symbol\":\"%v\",\"decimals\":%v}}]}", policyId, tokenAmount, symbol, decimals)
			log.Info(metaStr)
			ret.Metadata = make(map[string]interface{})
			err = json.Unmarshal([]byte(metaStr), &tokenBundleInst)
			if err != nil {
				log.Info(err.Error())
				return nil, err
			}
			metaDataInst.TokenBundle = append(metaDataInst.TokenBundle, tokenBundleInst)
		}
	}
	if len(metaDataInst.TokenBundle) > 0 {
		metaDataBytes, _ := json.Marshal(metaDataInst)
		json.Unmarshal(metaDataBytes, &ret.Metadata)
	}
	return ret, nil
}

func ToChangeOutputsOperation(toaddr string, Totaltokens map[string]uint64) (ret []*types.Operation, err error) {
	tokensBytes, _ := json.Marshal(Totaltokens)
	tokens := make(map[string]uint64)
	json.Unmarshal(tokensBytes, &tokens)
	adaMount := tokens["ADA"]
	delete(tokens, "ADA")
	if len(tokens) == 0 {
		if adaMount >= 1000000 {
			op, err := NewOutputOperation(toaddr, adaMount, "", 0)
			if err != nil {
				return nil, err
			}
			ret = append(ret, op)
			return ret, nil
		}
	}
	//num := len(tokens)
	//i := 1
	for k, v := range tokens {
		minada, ok := MinAdaMap[k]
		if !ok {
			minada = DefaultMinAda
		}
		if minada > adaMount {
			return nil, errors.New("出账不够minada")
		}

		//if i == num {
		//	minada = adaMount
		//}
		op, err := NewOutputOperation(toaddr, minada, k, v)
		if err != nil {
			log.Info(err.Error())
			return nil, err
		}
		ret = append(ret, op)
		adaMount -= minada
	}
	if adaMount >= 1000000 {
		op, err := NewOutputOperation(toaddr, adaMount, "", 0)
		if err != nil {
			return nil, err
		}
		ret = append(ret, op)
		return ret, nil
	}
	return ret, nil
}
