package v3

import (
	"github.com/gin-gonic/gin"
	dingModel "github.com/group-coldwallet/blockchains-go/model/dingding"
	"github.com/group-coldwallet/blockchains-go/pkg/httpresp"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
)

type TrxFixRequest struct {
	OrderNo string `json:"orderNo"`
}

//
//
//func TrxFix(c *gin.Context) {
//	var req TrxFixRequest
//	err := c.BindJSON(&req)
//	if err != nil {
//		httpresp.HttpRespErrWithMsg(c, err.Error())
//		return
//	}
//	dingRoleKey := dingding.DingUsers["$:LWCP_v1:$/sYvbj+f8dM4l3ALrh28Gw=="].RoleName
//	dingRole := dingding.DingRoles[dingRoleKey]
//	if dingRole == nil {
//		httpresp.HttpRespErrWithMsg(c, "缺少设置权限")
//		return
//	}
//	if err = abandonedOrder(dingModel.DING_ABANDONED_ORDER.ToString()+req.OrderNo+" ", dingRole); err != nil {
//		httpresp.HttpRespErrWithMsg(c, err.Error())
//		return
//	}
//	httpresp.HttpRespOkOnly(c)
//
//}

func TrxRePush(c *gin.Context) {
	var req TrxFixRequest
	err := c.BindJSON(&req)
	if err != nil {
		httpresp.HttpRespErrWithMsg(c, err.Error())
		return
	}
	dingRoleKey := dingding.DingUsers["$:LWCP_v1:$/sYvbj+f8dM4l3ALrh28Gw=="].RoleName
	dingRole := dingding.DingRoles[dingRoleKey]
	if dingRole == nil {
		httpresp.HttpRespErrWithMsg(c, "缺少设置权限")
		return
	}
	if err = DiscardAndRePush(dingModel.DING_DISCARD_REPUSH_ORDER.ToString()+req.OrderNo+" ", dingRole); err != nil {
		httpresp.HttpRespErrWithMsg(c, err.Error())
		return
	}
	httpresp.HttpRespOkOnly(c)

}
