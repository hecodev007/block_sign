package v3

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
)

var (
	defaultGasPrice = decimal.Zero
	addGasPrice     = decimal.NewFromInt(20)
)

//获取当前链上预估手续费
func GetEthEstFee(ctx *gin.Context) {
	var (
		tokenLimit  = decimal.NewFromInt(90000)
		ethLimit    = decimal.NewFromInt(30000)
		gasResult   = new(model.EthScanGasDataResult)
		ethFee      = decimal.Zero
		ethTokenFee = decimal.Zero
	)

	//获取线上eth gas
	url := conf.Cfg.EthScanCfg.Host + "/api?module=gastracker&action=gasoracle&apikey=" + conf.Cfg.EthScanCfg.Token
	resp, err := util.Get(url)
	if err != nil {
		httpresp.HttpRespCodeError(ctx, httpresp.STATUS_ERROR, err.Error(), nil)
		return
	}
	err = json.Unmarshal(resp, gasResult)
	log.Infof("eth gas请求结果:%s", string(resp))
	if err != nil {
		httpresp.HttpRespCodeError(ctx, httpresp.STATUS_ERROR, err.Error(), nil)
		return
	}
	if gasResult.Status != "1" || gasResult.Result == nil {
		httpresp.HttpRespCodeError(ctx, httpresp.STATUS_ERROR, string(resp), nil)
		return
	}
	//默认加10
	useGasPrice := gasResult.Result.FastGasPrice.Add(addGasPrice).Shift(9)

	ethFee = useGasPrice.Mul(ethLimit).Shift(-18)
	ethTokenFee = useGasPrice.Mul(tokenLimit).Shift(-18)
	result := &model.EthFeeResult{
		EthFee:      ethFee.String(),
		EthTokenFee: ethTokenFee.String(),
	}
	httpresp.HttpRespOK(ctx, httpresp.GetMsg(httpresp.SUCCESS), result)
}
