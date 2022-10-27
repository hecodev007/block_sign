package server

import (
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/api"
	"net/http"
)

func makeRouter() *http.ServeMux {
	router := http.NewServeMux()
	//测试接口
	router.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "Are you ok?")
	})

	//===========================使用的规范接口===========================
	////创建交易
	//router.HandleFunc("/create", api.CreateTransfer)
	////签名交易
	//router.HandleFunc("/sign", api.SignTransfer)
	////推送交易
	//router.HandleFunc("/push", api.PushTransfer)
	//===========================使用的规范接口===========================

	//===========================V1版本规范接口===========================
	//新版构建交易（多条）
	router.HandleFunc("/create", api.GetSignInputNew)
	//新版构建交易（单条）
	router.HandleFunc("/createOne", api.GetSignInputNewOne)
	//新版交易签名（多条）
	router.HandleFunc("/sign", api.SignNew)
	//新版交易签名（多条）
	router.HandleFunc("/signOne", api.SignNewOne)
	//单条广播
	router.HandleFunc("/push", api.PushTransaction)
	//新版广播交易 (多条)
	router.HandleFunc("/pushs", api.PushTransactionMore)
	//导入地址到客户端监控
	router.HandleFunc("/importaddrs", api.ImportAddress)

	//utxo打散接口
	//router.HandleFunc("/importaddrs", api.ImportAddress)
	//===========================V1版本规范接口===========================

	//===========================辅助接口===========================
	router.HandleFunc("/gas", api.Gas)
	//创建交易并且广播
	router.HandleFunc("/transaction_create", api.TransactionCreate)
	//获取交易utxo
	router.HandleFunc("/get_txin", api.GetTxInput)
	router.HandleFunc("/get_txin_fee", api.GetTxInputUseFee)
	//构建交易
	router.HandleFunc("/get_signinput", api.GetSignInput) //doc 1.0
	//交易签名
	router.HandleFunc("/sign_old", api.Sign) //doc 1.0
	//广播交易
	router.HandleFunc("/push_tx", api.PushTransaction) //doc 1.0
	//多签备用（尚未测试）
	//router.HandleFunc("/sign2", api.Sign2)
	//单线程创建地址
	router.HandleFunc("/create_address", api.BatchCreateAddressBySingleThread)
	//多线程批量生成地址，数量过大目前存在bug，暂时弃用
	//router.HandleFunc("/batch_create_address", api.BatchCreateAddress)
	//上传AB文件,解密在内存
	router.HandleFunc("/upload", api.Upload)
	//获取私钥
	router.HandleFunc("/get_key", api.GetPrivateKey)
	//导入私钥并且导入到rpc客户端
	router.HandleFunc("/import_key", api.ImportPrivKey)
	//导入私钥,不导入到rpc客户端
	router.HandleFunc("/import_key2", api.ImportPrivKey2)
	//旧项目兼容导入，明文传入
	router.HandleFunc("/import_key3", api.ImportPrivKey2)
	//删除内存中的key
	router.HandleFunc("/remove_key", api.RemovePrivKey)
	//上传TransactionInput文件批量生成签名参数
	router.HandleFunc("/get_signinputs", api.UploadTransactionInput)
	//上传SignInput文件批量生成签名交易
	router.HandleFunc("/signs", api.UploadSignInput)
	//上传PushInput文件批量广播交易
	router.HandleFunc("/push_transactions", api.UploadPushInput)
	//下载文件
	router.HandleFunc("/download/", api.DownLoad)
	//===========================辅助接口===========================
	return router
}
