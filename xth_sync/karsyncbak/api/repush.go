package api

import (
	"github.com/gin-gonic/gin"
	"karsync/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18899/yotta/repush -d '{"uid":1,"txid":"8e25973a4c85d0089a0551e649fd2dfe9a05da759c1ba0578b5c7c1414053d3f"}'
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
