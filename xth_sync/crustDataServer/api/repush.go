package api

import (
	"crustDataServer/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18893/xrp/repush -d '{"uid":1,"txid":"09F19FD2159E9882B30F0EB3057F94FA5475B7D931F55B8768707F706B487E99"}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error()+" need:uid=xxx&txid=tttt&height")
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Txid, req.Height); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
