package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"ytaDataServer/models/bo"
)

// @Description insert watch address
// @Accept  json
// @Produce  json
// @Param   addresses body  []models.InsertRequest  true   "add addresses"
func (m *MController) InsertWatchAddress(c *gin.Context) {
	var reqs []*bo.InsertRequest
	if err := c.BindJSON(&reqs); err != nil {
		NewError(c, err.Error())
		return
	} else if len(reqs) == 0 {
		NewError(c, fmt.Sprintf("no have reqs elm"))
		return
	}

	for _, r := range reqs {
		m.watcher.InsertWatchAddress(r.UserId, r.Address, r.Url)
	}
	NewSucc(c, "ok")
	return
}

// @Description insert watch contract
// @Accept  json
// @Produce  json
// @Param   addresses body  []bo.InsertContractRequest  true   "add contract"
func (m *MController) InsertWatchContract(c *gin.Context) {
	var reqs []*bo.InsertContractRequest
	if err := c.BindJSON(&reqs); err != nil {
		NewError(c, err.Error())
		return
	} else if len(reqs) == 0 {
		NewError(c, fmt.Sprintf("no have reqs elm"))
		return
	}

	for _, r := range reqs {
		m.watcher.InsertWatchContract(r.Name, r.ContractAddress, r.CoinType, r.Decimal)
	}

	NewSucc(c, "ok")
	return
}
