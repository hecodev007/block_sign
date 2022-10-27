package custody

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/service"
	"strings"
)

//CreateClientIdSecret 商户 client_id,安全secret
func CreateClientIdSecret(c *gin.Context) {

	name := c.PostForm("name")
	phone := c.PostForm("phone")
	email := c.PostForm("email")
	companyImg := c.PostForm("company_img") //企业图片
	req := model.RegisterMchRequest{
		Name:name,
		Phone:phone,
		Email:email,
		CompanyImg:companyImg,
	}
	newInfo, err := service.RegisterFcMchService(req)
	if err != nil {
		log.Errorf("RegisterFcMchService err = %v",err)
		if strings.Contains(err.Error(),"Duplicate entry") {
			eArr := strings.Split(err.Error()," ")
			if len(eArr) >= 8{
				httpresp.HttpRespError(c, httpresp.FAIL, fmt.Sprintf("注册商户数据重复： %v,%v", eArr[7], eArr[4]), nil)
				return
			}
		}
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("注册商户错误： %v", err.Error()).Error(), nil)
		return
	}
	httpresp.HttpRespOK(c, "success", newInfo)
}

//ReSecretClientIdSecret 重置商户密钥
func ReSecretClientIdSecret(c *gin.Context) {

	apikey := c.PostForm("api_key")
	req := model.RegisterMchRequest{
		ApiKey:apikey,
	}
	newInfo, err := service.ReSetFcMchSecretService(req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("重置商户密钥错误： %v", err.Error()).Error(), nil)
		return
	}
	httpresp.HttpRespOK(c, "success", newInfo)
}

//SearchClientIdSecret 查询商户密钥
func SearchClientIdSecret(c *gin.Context) {

	apikey := c.PostForm("api_key")
	req := model.RegisterMchRequest{
		ApiKey:apikey,
	}
	newInfo, err := service.SearchFcMchSecretService(req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("重置商户密钥错误： %v", err.Error()).Error(), nil)
		return
	}
	httpresp.HttpRespOK(c, "success", newInfo)
}

//BindAddress 绑定用户充值地址
func BindAddress(c *gin.Context) {

	apiKey := c.PostForm("api_key")
	address := c.PostForm("address")
	coinName := c.PostForm("coin_name") //逗号拼接的字符串
	wIp := c.PostForm("ip")
 var err error
	if apiKey == "" {
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Sprintf("商户Id错误"), nil)
		return
	}
	req := model.BindMchRequest{
		ApiKey:apiKey,
		Address:address,
		CoinName:coinName,
		WhiteIp: wIp,
	}
	err = service.BindAddressService(req)
	if err != nil {
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("商户绑定地址错误： %v", err.Error()).Error(), nil)
		return
	}
	httpresp.HttpRespOK(c, "success", nil)
}