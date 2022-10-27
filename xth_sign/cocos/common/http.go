package common

import (
	"bytes"
	"github.com/astaxie/beego/logs"
	"github.com/group-coldwallet/cocos/model/bo"
	"github.com/group-coldwallet/cocos/model/constants"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

func HttpRequest(method string, url string, payload []byte) *bo.HttpReturn {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	httpReturn := new(bo.HttpReturn)
	httpReturn.Code = 0
	httpReturn.Error = constants.Http200
	httpReturn.Body = ""
	req.Header.Set("Content-Type", "application/json")
	if viper.GetBool("rpc.enable") {
		req.Header.Add("Authorization", "Basic "+EncodeBasicAuth(viper.GetString("rpc.username"), viper.GetString("rpc.password")))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Debug(err.Error())
		httpReturn.Error = constants.HttpRequestError
		httpReturn.Code = constants.HttpRequestErrorCode
		return httpReturn
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	httpReturn.Body = string(body)
	return httpReturn
}
