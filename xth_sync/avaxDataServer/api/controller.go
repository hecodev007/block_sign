package api

import (
	"avaxDataServer/common"
	"avaxDataServer/services"
	"github.com/gin-gonic/gin"
	"net/http"
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

func NewError(ctx *gin.Context, msg string) {
	er := HTTPError{
		Code:    -1,
		Message: msg,
	}
	ctx.JSON(http.StatusOK, er)
}

func (s *MController) Info(c *gin.Context) {
	name, height, err := s.processor.Info()
	if err != nil {
		NewError(c, err.Error())
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
