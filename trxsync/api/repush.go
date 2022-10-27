package api

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/trxsync/models/bo"
	"net/http"
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
func (s *MController) RepushTx(c *gin.Context) {
	req := &bo.RePushRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	err := s.rePush(req.UserId, req.Txid, req.Height)
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

func (s *MController) rePush(userId int64, txid string, height int64) error {
	retry, err := s.bs.RepushTx(userId, txid, height)
	if err != nil {
		if retry {
			log.Infof("txId=%s 重推没有找到关心的地址，需要重新reload数据")
			if err = s.bs.Watcher.Reload(); err != nil {
				log.Errorf("watch Reload error %s", err.Error())
				return err
			}
			log.Infof("txId=%s watcher reload数据完成")
			_, err = s.bs.RepushTx(userId, txid, height)
		}
	}
	return err
}
