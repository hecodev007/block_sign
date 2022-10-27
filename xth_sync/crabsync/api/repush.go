package api

import (
	"crabsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15008/kar-kar/repush -d '{"uid":1,"txid":"0xaa52da479d93e4fc40f8a6c0dcbdddf12c8b8b7f519c1ae414161a9edc89138d","height":225428}'
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
