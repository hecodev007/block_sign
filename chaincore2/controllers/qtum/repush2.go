package qtum

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/astaxie/beego"
//	"github.com/group-coldwallet/chaincore2/common"
//	dao "github.com/group-coldwallet/chaincore2/dao/daoqtum"
//	"github.com/group-coldwallet/chaincore2/service/qtum"
//	"github.com/group-coldwallet/common/log"
//)
//
//type RepushTx2Controller struct {
//	beego.Controller
//}
//
//type RespData struct {
//	Result interface{}
//	Error 	interface{}
//	Code 	string
//}
//func rpcRequest(method string,params ...interface{})(interface{},error){
//	resp,err:=common.Request(method, []interface{}{params})
//	if err != nil {
//		return nil, fmt.Errorf("[%s] rpc error,err=%v",method,err)
//	}
//	var respD RespData
//	err = json.Unmarshal(resp,&respD)
//	if err != nil {
//		return nil, fmt.Errorf("[%s] json unmarshal error,err=%v",method,err)
//	}
//	if respD.Error!=nil {
//		return nil,fmt.Errorf("[%s] response error is not null,err=%v",respD.Error)
//	}
//	if respD.Result==nil {
//		return nil, fmt.Errorf("[%s] result is null",method)
//	}
//	return respD.Result, nil
//}
//
//func (c *RepushTx2Controller) Post() {
//	// 返回数据
//	resp := map[string]interface{}{
//		"code":    0,
//		"message": "ok",
//		"data":    nil,
//	}
//
//	set_resp := func(code int, message string) {
//		resp["code"] = code
//		resp["message"] = message
//	}
//
//	var jsonObj map[string]interface{}
//	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
//	log.Debug(jsonObj)
//	for true{
//		if jsonObj["txid"] == nil || jsonObj["uid"] == nil||jsonObj["height"]==nil {
//			set_resp(1, "param error")
//			break
//		}
//		// 读取 txid
//		txid := jsonObj["txid"].(string)
//		uid := int64(jsonObj["uid"].(float64))
//		height:=int64(jsonObj["height"].(float64))
//		if qtum.UserWatchList[uid] == nil {
//			set_resp(1, "user not found")
//			break
//		}
//		resp1,err:=rpcRequest("getblockhash",height)
//		if err != nil {
//			set_resp(1,fmt.Sprintf("%v",err))
//			break
//		}
//		blockHash:=resp1.(string)
//		resp2,err:=rpcRequest("getblock",blockHash,1)
//		if err != nil {
//			set_resp(1,fmt.Sprintf("%v",err))
//			break
//		}
//		result:=resp2.(map[string]interface{})
//
//		txs := result["tx"].([]interface{})
//		if len(txs)>0 {
//			for i,tx:=range txs{
//				if	tx.(string)==txid && i==1{
//					//这是一笔coinstake交易，不进行处理
//					set_resp(1,fmt.Sprintf("[%s] is a coinstake tx",txid))
//					return
//				}
//			}
//			//处理txid
//			resp3,err:=rpcRequest("getrawtransaction",txid,1)
//			if err != nil {
//				set_resp(1,fmt.Sprintf("%v",err))
//				break
//			}
//			blockInfo := qtum.Parse_block(result, false)
//			if blockInfo == nil {
//				set_resp(1,"parse block info error")
//				break
//			}
//
//		}
//
//	}
//
//}
//
