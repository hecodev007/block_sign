package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"dogeDataServer/models/bo"
	"net/http"
)

// @Description remove watch addresses
// @Accept  json
// @Produce  json
// @Param   addresses body  []models.RemoveRequest  true   "remove address"
func (m *MController) RemoveWatchAddress(c *gin.Context) {
	var reqs []*bo.RemoveRequest
	if err := c.BindJSON(reqs); err != nil {
		NewError(c, err.Error())
		return
	}

	if len(reqs) == 0 {
		NewError(c, fmt.Sprintf("no have reqs elm"))
		return
	}

	for _, r := range reqs {
		m.watcher.RemoveWatchAddress(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}

func (m *MController) RemoveWatchContract(c *gin.Context) {
	var reqs *bo.RemoveContractRequest
	if err := c.BindJSON(&reqs); err != nil {
		NewError(c, err.Error())
		return
	} else if len(reqs.ContractAddresses) == 0 {
		NewError(c, fmt.Sprintf("no have reqs elm"))
		return
	}

	res := ""
	for _, contract := range reqs.ContractAddresses {
		if err := m.watcher.RemoveWatchContract(contract); err != nil {
			res = res + err.Error() + ","
		}
	}
	NewSucc(c, res)
	return
}
