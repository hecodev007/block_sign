package api

import (
	"lunasync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//:4741325,
//curl -X POST --url http://127.0.0.1:15032/luna/repush -d '{"uid":1,"txid":"6B2D28831FB578E7072F5F73B3C5A2B8AB85EFC5B5A872DB537D9F5B1B0995EA"}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	if err := m.processor.RepushTx(req.UserId, req.Txid); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
