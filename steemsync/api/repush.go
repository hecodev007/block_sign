package api

import (
	"steemsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15009/eth/repush -d '{"uid":1,"txid":"0xab734d87efaeaf0975a88d915b95dfa4d0c2fae6059e0e3a8128806306130745"}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}

//curl -X POST --url http://127.0.0.1:15056/steem/repush -d '{"uid":3,"txid":"47ebbcc58183804521be6eb320ecfd6be31f2cba"}'
