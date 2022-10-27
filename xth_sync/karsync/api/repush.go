package api

import (
	"karsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15008/kar/repush -d '{"uid":1,"txid":"0x77af1a9529ca3a532e159ff0c1fbe2c42e121ab4e7ae86fd0c478d2c73ba5656","height":1739073}'
//curl -X POST --url http://127.0.0.1:15008/kar/repush -d '{"uid":1,"txid":"0x045cbb7c3d4c348ab71e7734c85fdcd5ad41f4327055a4ad506604d0ab38975e","height":1454733}'
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
