package api

import (
	"filDataServer/models/bo"
	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
// 更新下一个块的cids
//curl -X POST --url http://127.0.0.1:18896/fil/repush -d '{"uid":1,"txid":"bafy2bzaceagt6sk7kf6vj44izl76ihgbpahab77m4yo7cgobzlyompt75f7oa","height": 1600313}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Txid, req.Height); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
