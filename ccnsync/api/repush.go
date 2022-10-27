package api

import (
	"encoding/json"
	"log"
	"solsync/models/bo"

	"github.com/gin-gonic/gin"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
//curl -X POST --url http://127.0.0.1:15009/okt/repush -d '{"uid":1,"txid":"0xadc3c690268e6465b9987f1add21d26fa71e5974404f59cdf6ee36ee21a42a6c"}'
func (m *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		NewError(c, err.Error())
		return
	}

	dd, _ := json.Marshal(req)
	log.Printf("RepushTx req data======>:%s \n", string(dd))

	//if req.Height == 0 {
	//	c.JSON(http.StatusOK, gin.H{
	//		"code":    -1,
	//		"message": "需要添加高度(height)补数据",
	//	})
	//	return
	//}

	if err := m.processor.RepushTx(req.UserId, req.Txid); err != nil {
		NewError(c, err.Error())
		return
	}

	NewSucc(c, "ok")
	return
}
