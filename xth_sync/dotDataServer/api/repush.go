package api

import (
	"github.com/gin-gonic/gin"
	"dotDataServer/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18876/dot/repush -d '{"uid":1,"txid":"0x49fc40b60f244c00a85997c7ad1f7d9542a6193a7f0f5571f2762a532d23cc7d","height":6230938}'
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
