package api

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/scanning-service/services"
	"net/http"
)

type MController struct {
	bs *services.BaseService
}

func NewMController(bs *services.BaseService) (*MController, error) {
	return &MController{
		bs: bs,
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
	name, height, err := s.bs.Info()
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
		"data": ResInfo{
			Coin:   name,
			Height: height,
		},
	})
	return
}

type ResInfo struct {
	Coin   string `json:"coin"`
	Height int64  `json:"height"`
}
