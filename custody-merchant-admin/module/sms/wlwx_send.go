package sms

import (
	"fmt"
	wsms "github.com/wlwx/go-sdk/sms"
)

type WlwxEmailConfig struct {
	CustomName   string `json:"custom_name"`
	CustomPwd    string `json:"custom_pwd"`
	SmsClientUrl string `json:"sms_client_url"`
	Uid          string `json:"uid"`
	Content      string `json:"content"`
	DestMobiles  string `json:"cest_mobiles"`
	NeedReport   bool   `json:"need_report"`
	SpCode       string `json:"sp_code"`
}

func (w *WlwxEmailConfig) SendWlwxSms() (bool, error) {
	custom_name := w.CustomName
	custom_pwd := w.CustomPwd
	fmt.Printf("发送内容：%s，手机号码：%s \n", w.Content, w.DestMobiles)
	sms_client := wsms.NewSmsClient(w.SmsClientUrl, custom_name, custom_pwd)
	// 发送普通短信（业务标识uid选填）
	req := &wsms.SmsReq{
		Content:     w.Content,
		DestMobiles: w.DestMobiles,
		Uid:         w.Uid,
		NeedReport:  w.NeedReport,
		SpCode:      w.SpCode,
		MsgFmt:      wsms.SmsMsgUCS2,
	}
	resp, err := sms_client.SendMsg(req)

	//ss := &wsms.SmsVeriantInput{
	//	Content: "${mobile}用户您好，今天{$var1}的天气，晴，温度${var2}度，事宜外出。",
	//	Uid:     "1",
	//	Params: []*wsms.MobileVars{
	//		&wsms.MobileVars{
	//			Mobile: "18707873353",
	//			Vars:   []string{"18707873353", "阴天", "11"},
	//		},
	//	},
	//	NeedReport: true,
	//	SpCode:     "",
	//	MsgFmt:     wsms.SmsMsgUCS2,
	//}
	//msg, err := sms_client.SendVariantMsg(ss)
	//if err != nil {
	//	return false, err
	//}

	//fmt.Printf("Resp:%v\n", msg)
	// 获取Token
	//token, err := sms_client.GetToken()
	//if err != nil {
	//	return false,err
	//}
	// 获取用户上行
	//mo, err := sms_client.GetMO()
	//if err != nil {
	//	return false, err
	//}
	// 获取状态报告
	//report, err := sms_client.GetReport()
	//if err != nil {
	//	return false, err
	//}
	// 获取账户余额
	//account, err := sms_client.QueryAccount()
	//if err != nil {
	//	return false, err
	//}

	if err != nil {
		fmt.Printf("Error:%s\n", err.Error())
		return false, err
	} else {
		fmt.Printf("Resp:%v\n", resp)
		return true, err
	}
}
