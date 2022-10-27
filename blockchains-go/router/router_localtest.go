package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/middleware"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"io/ioutil"
)

//用于本地测试接口

func InitTestRouter(r *gin.Engine) {
	//回调测试
	r.POST("/v1/callback", CallBack)
	r.POST("/test", middleware.CheckApiSign(), middleware.AuthApi(), CallBack)
}
func CallBack(ctx *gin.Context) {
	bodyByte, _ := ioutil.ReadAll(ctx.Request.Body)
	log.Infof("接收回调的内容是：%s", string(bodyByte))
	//cb := new(model.NotifyOrderToMch)
	//err := json.Unmarshal(bodyByte, cb)
	//if err != nil {
	//	httpresp.HttpRespCodeErrOnly(ctx)
	//	return
	//}
	httpresp.HttpRespCodeOkOnly(ctx)
}
