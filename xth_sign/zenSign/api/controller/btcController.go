package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type BtcController struct {
}

func (this *BtcController) Router(r *gin.Engine) {
	group := r.Group("/v1/btc")
	{
		group.POST("/createaddr", this.createAddress)
		group.POST("/sign", this.sign)
	}
}

func (this *BtcController) NewError(ctx *gin.Context, status int, err string) {
	ctx.JSON(status, gin.H{
		"code":    status,
		"message": err,
	})
}

func (this *BtcController) createAddress(ctx *gin.Context) {
	this.NewError(ctx, http.StatusOK, "test")
	return
}
func (this *BtcController) sign(ctx *gin.Context) {
	this.NewError(ctx, http.StatusOK, "test")
	return
}
