package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"github.com/group-coldwallet/scanning-service/models/bo"
)

// @Description insert watch address
// @Accept  json
// @Produce  json
// @Param   addresses body  []models.InsertRequest  true   "add addresses"
// @Success 200 {object}   HTTPError
// @Failure 400 {object}   HTTPError
// @Failure 404 {object}   HTTPError
// @Failure 500 {object}   HTTPError
// @Router /xxx/insert [post]
func (s *MController) InsertWatchAddress(c *gin.Context) {
	var reqs []*bo.InsertRequest
	if err := c.BindJSON(&reqs); err != nil {
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
		s.bs.Watcher.InsertWatchAddress(r.UserId, r.Address, r.Url)
	}
	log.Infof("当前内存监听地址个数为：%d", s.bs.Watcher.GetWatchAddressNums())
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}

// @Description insert watch contract
// @Accept  json
// @Produce  json
// @Param   addresses body  []bo.InsertContractRequest  true   "add contract"
// @Success 200 {object}   HTTPError
// @Failure 400 {object}   HTTPError
// @Failure 404 {object}   HTTPError
// @Failure 500 {object}   HTTPError
// @Router /xxx/insertcontract [post]
func (s *MController) InsertWatchContract(c *gin.Context) {
	var reqs []*bo.InsertContractRequest
	if err := c.BindJSON(&reqs); err != nil {
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
		s.bs.Watcher.InsertWatchContract(r.Name, r.ContractAddress, r.CoinType, r.Decimal)
	}
	log.Infof("当前监听合约地址数量： %d", s.bs.Watcher.GetWatchContractNums())
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
