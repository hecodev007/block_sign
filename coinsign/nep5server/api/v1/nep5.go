package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/group-coldwallet/nep5server/conf"
	"github.com/group-coldwallet/nep5server/internal/common"
	"github.com/group-coldwallet/nep5server/model/bo"
	"github.com/group-coldwallet/nep5server/model/vo"
	"github.com/group-coldwallet/nep5server/service"
	"github.com/group-coldwallet/nep5server/service/nep5service"
)

type Nep5API struct {
	config        *conf.Config
	transfService service.TransfService
}

func NewNep5API(cfg *conf.Config) *Nep5API {
	return &Nep5API{
		config:        cfg,
		transfService: nep5service.NewNep5Service(),
	}
}

func (v1api *Nep5API) CreateAddr(c *gin.Context) {
	params := new(bo.CreateAddrBO)
	err := c.ShouldBindBodyWith(params, binding.JSON)
	if err != nil {
		common.HttpRespError(c, err.Error(), common.ERROR_JSON_BODY)
		return
	}
	if params.CoinName != "nep5" {
		common.HttpRespError(c, common.GetMsg(common.ERROR_COIN), common.ERROR_COIN)
		return
	}
	params.CreatePath = v1api.config.PemPath
	result, err := v1api.transfService.CreateAddr(params)
	if err != nil {
		common.HttpRespError(c, common.GetMsg(common.ERROR), common.ERROR)
		return
	}
	common.HttpRespOK(c, result)
}

func (v1api *Nep5API) Transf(c *gin.Context) {
	params := new(bo.TransfBO)
	err := c.ShouldBindBodyWith(params, binding.JSON)
	if err != nil {
		common.HttpRespError(c, err.Error(), common.ERROR_JSON_BODY)
		return
	}
	checkResult := false
	for _, v := range v1api.config.Nep5Cfg {
		if params.ScriptAddr == v.AssetsId {
			//金额对比
			println(params.AmountFloatStr.Shift(v.Decimal).IntPart() == params.AmountInt)
			if params.AmountFloatStr.Shift(v.Decimal).IntPart() == params.AmountInt {
				checkResult = true
			}
			break
		}
	}
	if !checkResult {
		common.HttpRespError(c, common.GetMsg(common.ERROR_AMOUNT), common.ERROR_AMOUNT)
		return
	}
	raw, txid, err := v1api.transfService.TokenSign(params.From, params.To, params.ScriptAddr, params.AmountInt)
	if err != nil {
		common.HttpRespError(c, common.GetMsg(common.ERROR), common.ERROR)
		return
	}
	result := vo.TransfResultVO{
		Hex:  raw,
		Txid: txid,
	}
	common.HttpRespOK(c, result)

}
