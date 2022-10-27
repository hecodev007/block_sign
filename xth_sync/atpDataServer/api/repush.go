package api

import (
	"atpDataServer/models/bo"
	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:18903/atp/repush -d '{"uid":1,"txid":"0xebdd67cdb8d56d9a0fff9db2b29c40851bd33e564cfbaa0a3249123883213247"}'
//curl -X POST --url http://127.0.0.1:18903/atp/repush -d '{"uid":1,"txid":"0xd7086bcbfcb7f37723c71312772c7a1d67af92d6c1369c90e4cad3859452e92c"}'

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

	NewSucc(c, "ok;注,以后atp代币假充值补推送,要人工确认成功,")
	return
}
