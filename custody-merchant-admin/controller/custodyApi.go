package controller

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/router/web/handler"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func PushPassOrPushBill(c *handler.Context) error {
	// 正确处理
	data, _ := ioutil.ReadAll(c.Request().Body)
	mp := map[string]interface{}{}
	err := json.Unmarshal(data, &mp)
	if err != nil {
		fmt.Println(err.Error())
		log.Error(err.Error())
		return handler.NewError(c, err.Error())
	}
	msg := ""
	if v, ok := mp["msg"]; ok {
		msg = v.(string)
		fmt.Printf("%s,data:%v", msg, string(data))
		log.Infof("RPC接收消息 %s,data：%v", msg, string(data))
	} else {
		fmt.Println("缺少msg参数")
		log.Error("缺少msg参数")
		return handler.NewError(c, err.Error())
	}

	if v, ok := mp["type"]; ok {
		billData := domain.BillInfo{}
		b, err := json.Marshal(mp["data"])
		if err != nil {
			log.Error(err.Error())
			return handler.NewError(c, err.Error())
		}
		json.Unmarshal(b, &billData)
		params := mp["params"].(map[string]interface{})
		switch v {
		case "re_push":
			// TODO 重推,这里写商户回调逻辑
			if billData.SerialNo == "" {
				log.Error("重推失败，订单号 为null")
				return handler.NewError(c, "重推失败，serialNo 为null")
			}
			err = service.PushDataByUrl(billData.SerialNo)
			if err != nil {
				log.Error(err.Error())
				return handler.NewError(c, err.Error())
			}
			break
		case "withdrawal":
			// TODO 提现,这里写提现审核逻辑
			billStatus := params["bill_status"].(float64)
			err = service.MerchantWithdrawal(billData, int(billStatus))
			if err != nil {
				return handler.NewError(c, err.Error())
			}
		}
	}
	res := handler.NewResult(0, "")
	return res.ResultOk(c)
}
