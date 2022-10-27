package api

import (
	"nyzoDataServer/models/bo"
	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:22011/nyzo/repush -d '{"uid":1,"txid":"ae4a235a129e88ba7dd7e78780bad41075ec4bd2a2a2982962e5145374ca1624c60c3a867f658059934d6eb19dd71d87633287909cd5605e1e012e3761409807","height":11395456}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Txid,req.Height); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
