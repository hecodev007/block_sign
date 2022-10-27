package controller

import (
	"bosSign/api/models"
	"bosSign/common/log"
	"bosSign/common/validator"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/eoscanada/eos-go/ecc"

	eos "github.com/eoscanada/eos-go"
	"github.com/gin-gonic/gin"
)

type EosController struct {
	mod models.EosModel
}

func (this *EosController) Router(r *gin.Engine) {
	r.POST("/sign", this.coldSign)
}
func (this *EosController) NewError(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": err,
	})
}

func (this *EosController) coldSign(ctx *gin.Context) {
	var params = new(validator.ColdSign)
	if err := ctx.ShouldBindJSON(params); err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	log.Info("/sign and get params:", String(params))
	chainid, _ := hex.DecodeString(params.ChainID)
	payload, _ := hex.DecodeString(params.Data)
	//hash, _ := hex.DecodeString(params.Hash)
	//t.Log(string(hash))
	digest := eos.SigDigest(chainid, payload, []byte{})
	wifKey, err := new(models.EosModel).GetPrivate(params.MchID, params.PublicKey)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	privKey, err := ecc.NewPrivateKey(string(wifKey))
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	sigrature, err := privKey.Sign(digest)
	if err != nil {
		this.NewError(ctx, err.Error())
		return
	}
	result := new(validator.ColdSignResult)
	result.Data.ColdData = params.ColdData
	result.Message = "ok"
	result.Data.Signatures = append(result.Data.Signatures, sigrature.String())
	ctx.JSON(http.StatusOK, result)
	log.Info("return success:", String(result))
	return
}

func String(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}
