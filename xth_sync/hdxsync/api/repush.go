package api

import (
	"github.com/gin-gonic/gin"
	"hdxsync/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18899/dot/repush -d '{"uid":1,"txid":"0xfe7dac4bb30c256d07964ea45c8790aa001c8effa480c5ee734e8361e3c58800","height":6230938}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId,req.Height, req.Txid); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
