package global

//
////临时记录 用于跳过并发的账号模型请求
////储存 商户ID 对应的币种执行列表的订单 如： 1 btc btc0000001
//var AccountTransfer map[int][]*CoinOrderInfo
//
//type CoinOrderInfo struct {
//	CoinName   string `json:"coinName"`
//	OutOrderNo string `json:"outOrderNo"`
//}
//
//func init() {
//	//AccountTransfer = make(map[int64][]*CoinOrderInfo, 0)
//	AccountTransfer = make(map[int][]*CoinOrderInfo, 0)
//}
//
////查询临时拦截变量是否存在交易
//func HasAccountTransfer(appid int, coinName string) (ok bool, outOrderNo string) {
//	sliceCoin := AccountTransfer[appid]
//	for _, v := range sliceCoin {
//		if v.CoinName == coinName {
//			return true, v.OutOrderNo
//		}
//	}
//	return false, ""
//}
//
////添加临时拦截
//func AddAccountTransfer(appid int, coinName, outOrderNo string) bool {
//	sliceCoin := AccountTransfer[appid]
//	if sliceCoin == nil {
//		sliceCoin = make([]*CoinOrderInfo, 0)
//	}
//	sliceCoin = append(sliceCoin, &CoinOrderInfo{
//		CoinName:   coinName,
//		OutOrderNo: outOrderNo,
//	})
//	AccountTransfer[appid] = sliceCoin
//	return true
//}
//
////删除拦截
//func DelAccountTransfer(appid int, coinName string) bool {
//	sliceCoin := AccountTransfer[appid]
//	if sliceCoin == nil {
//		sliceCoin = make([]*CoinOrderInfo, 0)
//		return true
//	}
//	newsliceCoin := make([]*CoinOrderInfo, 0)
//	for i, v := range sliceCoin {
//		if v.CoinName != coinName {
//			newsliceCoin = append(newsliceCoin, sliceCoin[i])
//		}
//	}
//	AccountTransfer[appid] = newsliceCoin
//	return true
//}
