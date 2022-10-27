package router

import (
	"github.com/gin-gonic/gin"
	"github.com/group-coldwallet/blockchains-go/router/api"
	"github.com/group-coldwallet/blockchains-go/router/api/admin"
	v1 "github.com/group-coldwallet/blockchains-go/router/api/v1"
	"github.com/group-coldwallet/blockchains-go/runtime/wechat"
)

//基本不会有变动的路由
func InitAdminRouter(r *gin.Engine) {

	//删除定时任务
	r.POST("/admin/jobrunner/remove", gin.BasicAuth(gin.Accounts{
		"rylink": "hoo123!@#", //用户名：密码
	}), api.JobRemove)

	//刷新所有全局配置
	r.POST("/admin/refresh", gin.BasicAuth(gin.Accounts{
		"rylink": "hoo123!@#", //用户名：密码
	}), api.RefreshGlobal)

	//主动推送订单
	r.POST("/admin/pushorder", gin.BasicAuth(gin.Accounts{
		"rylink": "hoo123!@#", //用户名：密码
	}), v1.TransPushOrder)

	//修正订单，通常出现在eos订单，order表完成订单，但是业务订单未完成，应付walletserver回调失败状况
	r.POST("/admin/repairorder", gin.BasicAuth(gin.Accounts{
		"rylink": "hoo123!@#", //用户名：密码
	}), admin.RepairOrder)

	//测试wx
	r.POST("/admin/wx", gin.BasicAuth(gin.Accounts{
		"rylink": "hoo123!@#", //用户名：密码
	}), func(context *gin.Context) {
		wechat.SendWarnInfo("手动测试读取")
	})

}
