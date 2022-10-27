package api

import (
	"bncsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15021/bnc/repush -d '{"uid":1,"txid":"0x4b3d2267b1718311f9c9b97c4ccef2060b0db8e4d5c9446b8f250445f94a90f9","height":1545817}'
//curl -X POST --url http://127.0.0.1:15021/bnc/repush -d '{"uid":1,"txid":"0x809f5150e2a7fa1ca959872d1de07ed96d829e841372e9b41bf235f96a07e430","height":1544693}'
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
