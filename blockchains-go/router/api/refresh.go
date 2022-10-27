package api

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/runtime"
)

//刷新全局变量
func RefreshGlobal(c *gin.Context) {
	runtime.InitGlobalReload()
	httpresp.HttpRespCodeOkOnly(c)
}
