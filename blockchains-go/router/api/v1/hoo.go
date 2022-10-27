package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/service"
	"net/http"
)

func ApplyTransactionSubmit(c *gin.Context) {
	var (
		err     error
		appId   int64
		request = &model.TransferRequest{}
	)

	if err = c.BindJSON(request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  -1,
			"msg":     err.Error(),
			"success": false,
		})
		return
	}

	log.Debugf("交易申请,请求数据: %v", request)
	//判断是否为支持的币种
	appId, err = service.HooSvr.ApplyTransactionSubmit(request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  appId,
			"msg":     err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  appId,
		"msg":     "操作成功",
		"success": true,
	})
}
