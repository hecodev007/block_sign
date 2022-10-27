package api

import (
	"xlmDataServer/models/bo"
	"fmt"
	"github.com/gin-gonic/gin"
)

// @Description update watch address
// @Accept  json
// @Produce  json
// @Param   addresses body []models.UpdateRequest  true   "update address"
func (m *MController) UpdateWatchAddress(c *gin.Context) {
	var reqs []*bo.UpdateRequest
	if err := c.BindJSON(reqs); err != nil {
		NewError(c, err.Error())
		return
	} else if len(reqs) == 0 {
		NewError(c, fmt.Sprintf("no have reqs elm"))
		return
	}

	for _, r := range reqs {
		m.watcher.UpdateWatchAddress(r)
	}
	NewSucc(c, "ok")
	return
}
