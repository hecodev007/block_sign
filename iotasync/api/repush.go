package api

import (
	"github.com/gin-gonic/gin"
	"iotasync/common/log"
	"iotasync/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18894/iota/repush -d '{"uid":5,"txid":"362ffff65491d4c71958e27958ce9f069c26d102eb3b22e75acfb47d2f4b05c2"}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Txid); err != nil {
		log.Error(err.Error())
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
