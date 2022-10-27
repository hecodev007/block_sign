package v1

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/repush"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	v3 "github.com/group-coldwallet/blockchains-go/router/api/v3"
)

type rePushRequest struct {
	TxId   string `json:"txId"`
	Chain  string `json:"chain"`
	Height uint   `json:"height"`
}

func RePushTx(ctx *gin.Context) {
	params := rePushRequest{}
	ctx.BindJSON(&params)
	log.Infof("收到从checktx发送的补数据请求: %v", params)
	req := repush.DingRepush{
		Txid:   params.TxId,
		Coin:   params.Chain,
		Height: params.Height,
		Uid:    5,
	}

	ms, _ := json.Marshal(req)
	if err := v3.RepushChainData(string(ms), "checktx"); err != nil {
		httpresp.HttpRespErrWithMsg(ctx, err.Error())
		return
	}
	httpresp.HttpRespOK(ctx, "", nil)
}
