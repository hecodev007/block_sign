package api

import (
	"fmt"
	"lunasync/common"
	"lunasync/services"
	"net/http"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

type MController struct {
	processor common.Processor
	watcher   services.WatchControl
}

func NewMController(p common.Processor, w services.WatchControl) *MController {
	return &MController{
		processor: p,
		watcher:   w,
	}
}

func (m *MController) Router(r *gin.Engine, name string) {
	group := r.Group(fmt.Sprintf("/%s", name))
	{
		group.POST("/rpc", m.RpcPost)
		group.POST("/insert", m.InsertWatchAddress)
		group.POST("/remove", m.RemoveWatchAddress)
		group.POST("/update", m.UpdateWatchAddress)
		group.POST("/repush", m.RepushTx)
		group.POST("/insertcontract", m.InsertWatchContract)
		group.POST("/removecontract", m.RemoveWatchContract)
	}
	r.GET("/info", m.Info)
	//pprof  参考：github.com/DeanThompson/ginpprof
	ginpprof.Wrap(r)
}

func (m *MController) Info(c *gin.Context) {
	name, height, err := m.processor.Info()
	if err != nil {
		NewError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"coin":   name,
			"height": height,
		},
	})
	return
}

func NewError(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": msg,
	})
}

func NewSucc(ctx *gin.Context, msg string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": msg,
	})
}
