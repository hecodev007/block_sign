package custody

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"strconv"
)

func GetCoinList(c *gin.Context) {

	//a := c.Param("body.limit")
	//b := c.Param("body.offset")
	l,_ := strconv.Atoi(c.PostForm("limit"))
	o,_ := strconv.Atoi(c.PostForm("offset"))
	//
	//o := c.PostForm("offset")
	log.Infof("GetCoinList :%v,%v",l,o)
	result, err := api.CoinService.CustodyGetCoinList(l,o)
	if err != nil {
		log.Errorf("GetCoinList error:%s", err.Error())
		httpresp.HttpRespError(c, httpresp.FAIL, fmt.Errorf("获取币种列表失败： %v", err.Error()).Error(), nil)
		return
	}
	back := map[string]interface{}{
		"list":result,
	}
	httpresp.HttpRespOK(c, httpresp.GetMsg(httpresp.SUCCESS), back)
}
