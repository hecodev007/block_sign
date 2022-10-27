package models

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"steemsign/common/validator"
	"steemsign/utils/rpc"
	"steemsign/utils/rpc/transactions"
	"steemsign/utils/rpc/transports/rpcclient"
	"steemsign/utils/rpc/types"
)

//from:https://bos.eosn.io/v1/chain/get_info
//mainet
//curl -X POST --url http://127.0.0.1:14056/steem/transfer -d '{"mch_name":"1","coin_name":"steem","from_address":"steemgoapi","to_address":"marjay","quantity":"0.002"}'
type SteemModel struct {
	Url    string
	JsFile string
}

//获取私钥
func (m *SteemModel) GetPrivate() (private string, err error) {
	//todo:注释
	//if pubkey == "EOS6VeUZo93nzcmhK3HfQaXBsiw9tsd6hPfU2QwS2adpYQqM9G2Rt" {
	//	return "5JFnwrLsvo6nmCRPQ2636U2zygHZ9nj2YrNHh5WrTyxC4vwJ9q7", nil
	//}、
	//marjay
	//5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS

	return "5J68u6fFVeMCW7TUjSKs7N9GsawnZ8Wb91ggcJkMu5cqsq7z8ww", nil
}

func (m *SteemModel) SignTx(txParams validator.SignParams_Data) (p interface{}, err error) {
	//host := "https://api.steemit.com"

	url := m.Url

	t := rpcclient.NewRpcClient(url)

	// Use the transport to get an RPC client.
	// Use the transport to get an RPC client.
	client, err := rpc.NewClient(t)
	if err != nil {
		return nil, err
	}

	// Get the props to get the head block number and ID
	// so that we can use that for the transaction.
	props, err := client.Database.GetDynamicGlobalProperties()
	if err != nil {
		return nil, err
	}

	// Prepare the transaction.
	refBlockPrefix, err := transactions.RefBlockPrefix(props.HeadBlockID)
	if err != nil {
		return nil, err
	}

	//jsfile := "./offline_signing/sign_offline_transfer.js"
	jsfile := m.JsFile
	fmt.Println("jsfile path:", jsfile)
	wif, _ := m.GetPrivate()
	//s := fmt.Sprintf("node /Users/dnsb01357/js/offline_signing/sign_offline_transfer.js 46030 2055430191 marjay tipu '0.001' 5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
	s := fmt.Sprintf("node %s %v %v %s %s '%v' %s", jsfile, transactions.RefBlockNum(props.HeadBlockNumber), refBlockPrefix, txParams.FromAddress, txParams.ToAddress, txParams.Quantity, wif)

	cmd := exec.Command("/bin/bash", "-c", s)

	//// 执行命令，并返回结果
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Output->", err)
		return nil, err
	}
	fmt.Println("exec output:", string(output))
	trans := &types.Transaction{}
	json.Unmarshal(output, trans)
	return trans, nil
}

func (m *SteemModel) Transfer(txParams validator.SignParams_Data) (err error, id string) {
	// Process flags.
	//host := "http://10.0.230.86:8090"
	//host := "https://api.steemit.com"

	url := m.Url
	//brew install automake
	//brew install libtool
	//speck 安装
	//$ ./autogen.sh
	//$ ./configure --enable-module-recovery
	//$ make
	//$ make check
	//$ sudo make install

	t := rpcclient.NewRpcClient(url)

	// Use the transport to get an RPC client.
	// Use the transport to get an RPC client.
	client, err := rpc.NewClient(t)
	if err != nil {
		return err, ""
	}

	trans, err := m.SignTx(txParams)
	if err != nil {
		fmt.Println("broad cast error:", err)
		return err, ""
	}
	resp, err := client.NetworkBroadcast.BroadcastTransactionSynchronous(trans.(*types.Transaction))
	if err != nil {
		fmt.Println("broad cast error:", err)
		return err, ""
	}
	fmt.Printf("%+v\n", resp.ID)

	// Success!
	return nil, resp.ID
}
