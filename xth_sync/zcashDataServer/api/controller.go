package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zcashDataServer/common"
	"zcashDataServer/services"
)

type MController struct {
	processor common.Processor
	watcher   services.WatchControl
}

func NewMController(p common.Processor, w services.WatchControl) (*MController, error) {
	return &MController{
		processor: p,
		watcher:   w,
	}, nil
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewError(ctx *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}

func (s *MController) Info(c *gin.Context) {
	name, height, err := s.processor.Info()
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
		"data": common.ResInfo{
			Coin:   name,
			Height: height,
		},
	})
	return
}
