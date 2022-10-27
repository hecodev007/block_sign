package api

import (
	"github.com/gin-gonic/gin"
	"starDataServer/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18901/star/repush -d '{"uid":1,"txid":"bafy2bzaceakhgq57vlocygniabbhsobojqecmhzwgy6tuofhuumfjjrvgmlyc","height":43888}'
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
