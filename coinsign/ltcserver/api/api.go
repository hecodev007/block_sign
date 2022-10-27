package api

import "github.com/gin-gonic/gin"

//公共接口定义
type API interface {
	//创建签名模板
	CreateTpl(c *gin.Context)

	//签名
	SignTx(c *gin.Context)

	//广播接口
	SendTx(c *gin.Context)

	//check privkey 测试获取私钥
	GetPrivkey(c *gin.Context)

	//创建地址
	CreateAddrs(c *gin.Context)

	//地址校验
	CheckAddr(c *gin.Context)
}
