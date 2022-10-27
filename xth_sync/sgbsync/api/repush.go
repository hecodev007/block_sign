package api

import (
	"sgbsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15005/sgb-sgb/repush -d '{"uid":1,"txid":"0xc6315914dd7d0f29af41b9897e501b22ab95aa3163507b736a1c5c07c447a568","height":349957}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Height, req.Txid); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
