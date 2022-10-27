package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/trxsync/models/bo"
	"net/http"
)

// @Description remove watch addresses
// @Accept  json
// @Produce  json
// @Param   addresses body  []models.RemoveRequest  true   "remove address"
// @Success 200 {object}   HTTPError
// @Failure 400 {object}   HTTPError
// @Failure 404 {object}   HTTPError
// @Failure 500 {object}   HTTPError
// @Router /xxx/remove [post]
func (s *MController) RemoveWatchAddress(c *gin.Context) {
	var reqs []*bo.RemoveRequest
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
		s.bs.Watcher.RemoveWatchAddress(r)
	}
	log.Infof("当前内存监听地址个数为：%d", s.bs.Watcher.GetWatchAddressNums())
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}

func (s *MController) RemoveWatchContract(c *gin.Context) {
	var reqs *bo.RemoveContractRequest
	if err := c.BindJSON(&reqs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	if len(reqs.ContractAddresses) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("no have reqs elm"),
		})
		return
	}

	res := ""
	for _, contract := range reqs.ContractAddresses {
		if err := s.bs.Watcher.RemoveWatchContract(contract); err != nil {
			res = res + err.Error() + ","
		}
	}
	log.Infof("当前监听合约地址数量： %d", s.bs.Watcher.GetWatchContractNums())
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": res,
	})
	return
}
