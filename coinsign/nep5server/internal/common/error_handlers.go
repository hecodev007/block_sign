package common

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//通用错误返回
func PageNotFound(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	logrus.Printf("NoRoute claims: %#v\n", claims)
	HttpRespCommon(c, PAGE_NOT_FOUND, GetMsg(PAGE_NOT_FOUND), nil)
}
