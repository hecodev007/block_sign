package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/trxsync/models/bo"
	"net/http"
)

// @Description update watch address
// @Accept  json
// @Produce  json
// @Param   addresses body []models.UpdateRequest  true   "update address"
// @Success 200 {object}   HTTPError
// @Failure 400 {object}   HTTPError
// @Failure 404 {object}   HTTPError
// @Failure 500 {object}   HTTPError
// @Router /xxx/repush [post]
func (s *MController) UpdateWatchAddress(c *gin.Context) {
	var reqs []*bo.UpdateRequest
	if err := c.BindJSON(reqs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("no have reqs elm"),
		})
		return
	}

	for _, r := range reqs {
		s.bs.Watcher.UpdateWatchAddress(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
