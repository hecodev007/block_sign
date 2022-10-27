package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"telosDataServer/models/bo"
)

// @Description replay push tx
// @Accept  json
// @Produce  json
// @Param   repush body models.RePushRequest  true   "replay push tx"
// @Success 200 {object}   HTTPError
// @Failure 400 {object}   HTTPError
// @Failure 404 {object}   HTTPError
// @Failure 500 {object}   HTTPError
// @Router /xxx/repush [post]
// eg:curl -X POST --url http://127.0.0.1:18890/telos/repush -d '{"uid":1,"txid":"5fc58ca973913ad0469a99c041ab7e019f8f38cf769858210526eb82ec5aa386"}'
func (s *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	err := s.processor.RepushTx(req.UserId, req.Txid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
