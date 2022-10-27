package api

import (
	"bifrostsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15008/kar/repush -d '{"uid":1,"txid":"0x8dde860ef4b646ab059b9d6dd2c5ddc03ee1bb4f4b5cae25a012b16594d57c2f","height":347003}'
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
