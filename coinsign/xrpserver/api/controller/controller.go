package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"xrpserver/api/model"
	"xrpserver/common/log"
	"xrpserver/pkg/xrputil"
)

type Controller struct {
}

func (m *Controller) Router(r *gin.Engine,name string){
	group := r.Group(fmt.Sprintf("/%s", name))
	{
		group.POST("/transfer", m.transfer)
	}
}
func (this *Controller) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    -1,
		"message": err,
		"data":    "",
	})
}
func (this *Controller) transfer(ctx *gin.Context) {
	var params = new(model.TransferParams)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info(String(params))

	if params.From == "" || params.Amount== "" {
		this.NewError(ctx, "参数(from,amount),不能为空")
		return
	}
	if amount,err := decimal.NewFromString(params.Amount);err != nil {
		log.Info(err.Error()+"0000")
		this.NewError(ctx, "参数(amount),为浮点字符串")
		return
	} else if amount.Exponent() < -6{
		this.NewError(ctx, "参数(amount),精度为6位小数")
		return
	}

	if txid,err := xrputil.Transfer(params.From,params.Amount);err != nil{
		log.Info(err.Error())
		this.NewError(ctx, err.Error())
		return
	 } else {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": fmt.Sprintf("txid:%s",txid),
		})
		return
	}

}

func String(d interface{})string{
	str,_ := json.Marshal(d)
	return string(str)
}