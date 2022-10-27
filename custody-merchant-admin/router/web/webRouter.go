package web

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/controller"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	cd "custody-merchant-admin/middleware/casbin"
	"custody-merchant-admin/middleware/opentracing"
	"custody-merchant-admin/middleware/verify"
	"custody-merchant-admin/module/auth"
	"custody-merchant-admin/module/cache"
	"custody-merchant-admin/module/session"
	handler2 "custody-merchant-admin/router/web/handler"
	"errors"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

type (
	HandlerFunc func(*handler2.Context) error
)

// Routers
// web路由设置
func Routers() *echo.Echo {
	// Echo instance
	e := echo.New()
	// Context自定义
	e.Use(handler2.NewContext())
	// Customization
	if Conf.ReleaseMode {
		e.Debug = false
	}
	e.Logger.SetPrefix("web")
	e.Logger.SetLevel(GetLogLvl())
	// Session
	e.Use(session.Session())
	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	// Body Limit 中间件用于设置请求体的最大长度，如果请求体的大小超过了限制值，则返回 "413 － Request Entity Too Large" 响应
	e.Use(mw.BodyLimit("8M"))
	e.Use(mw.TimeoutWithConfig(mw.TimeoutConfig{
		Timeout: 90 * time.Second,
	}))
	// 验证码、静态资源使用http.ServeContent()，与Gzip有冲突，Nginx报错，验证码无法访问
	e.Use(mw.GzipWithConfig(mw.GzipConfig{
		Level: 5,
	}))
	// OpenTracing
	if !Conf.Opentracing.Disable {
		e.Use(opentracing.OpenTracing("web"))
	}
	// Cache
	e.Use(cache.Cache())
	// 财务路由
	financeRouter(e)
	// MQ
	//e.Use(mq.NewMQQueue(mq.DefaultMQConfig))
	e.GET("/health", handler(handler2.HealthHandler))
	// 测试
	if Conf.Mod == "dev" {
		testServer(e)
	}
	custodyRPC(e)
	openApiV1(e)
	coinServer(e)
	blockChainCallBackServer(e)
	//e.Use(cd.CasbinHandler())
	// 中间件配置自定义类型JWT
	e.GET("/ws", handler(handler2.WsHandler))
	e.Static("/static", "static")
	loginServer(e)
	baseServer(e)
	personnelServer(e)
	merchantSubServer(e)
	merchantMainServer(e)
	merchantChainsServer(e)
	chainBillServer(e)
	billServer(e)
	auditServer(e)
	assetsServer(e)
	packageServer(e)
	businessServer(e)
	merchantServer(e)
	financeServer(e)
	incomeServer(e)
	businessOrderServer(e)
	return e
}

func testServer(e *echo.Echo) {

	r := e.Group("/test")
	{
		r.POST("/rolleInCome", handler(controller.BlockchainIncomeCallback))
		r.POST("/uploadFile", handler(controller.UpLoadFiles))
		r.GET("/list", handler(controller.GetMerchantChainList))
		r.POST("/updateCoin", handler(controller.UpdateCoin))
		r.GET("/get", handler(controller.GetMerchantChainInfo))
		r.POST("/update", handler(controller.UpdateMerchantChainInfo))
		r.POST("/update/state", handler(controller.FreezeOrThawMerchantChainInfo))
		r.POST("/delete", handler(controller.DelMerchantChainInfo))
		//财务收益户
		r.GET("/income/list", handler(controller.GetIncomeList))
		//链上订单查询
		r.GET("/chain/bill/list", handler(controller.FindChainBillList))
		//链上订单导出
		r.GET("/chain/bill/export", handler(controller.FindChainBillExport))
		//收益户导出
		r.POST("/income/export", handler(controller.ExportIncomeList))
		//收益户导出
		r.GET("/income/chart", handler(controller.IncomeChartInfo))
		//资产管理列表、数据图
		r.GET("/assets/list", handler(controller.GetAssetsList))
		//测试使用
		r.POST("/merchant/withdraw", handler(controller.CreateWithdraw))
		//财务审核通过/审核拒绝
		r.POST("/finance/agree/item", handler(controller.FinanceAgreeRefuse))
		//提现
		r.POST("/wallet/withdraw", handler(controller.WalletWithdraw))
		r.POST("/order/rollback", handler(controller.OrderRollBack)) //订单回退
		//提现回调测试
		r.POST("/wallet/rollbackWithdraw", handler(controller.BlockchainCallback))
		r.POST("/clear/redis", handler(controller.ClearRedis))
		r.POST("/get/redis", handler(controller.GetRedisByKey))
		// 商户账单查询
		r.GET("/searchBill", handler(controller.FindBillInfos))
		// 批量拉取地址
		r.POST("/batchAddress", handler(controller.GenerateBatchAddress))
		// 拉取单个地址
		r.GET("/singleAddress", handler(controller.GenerateUserAddress))
		// 商户资产查询
		r.GET("/findAssets", handler(controller.FindAssetsByCoinList))
	}
}

func custodyRPC(e *echo.Echo) {

	r := e.Group("/api/rpc/v1")
	// 商户发送给管理后台的消息
	r.POST("/pushMsg", handler(controller.PushPassOrPushBill))
}

func openApiV1(e *echo.Echo) {
	r := e.Group("/api/open/v1")
	// 商户发起提现
	r.POST("/withdraw", handler(controller.CreateWithdraw))
	// 商户账单查询
	r.GET("/searchBill", handler(controller.FindBillInfos))
	// 批量拉取地址
	r.POST("/batchAddress", handler(controller.GenerateBatchAddress))
	// 拉取单个地址
	r.GET("/singleAddress", handler(controller.GenerateUserAddress))
	// 商户资产查询
	r.GET("/findAssets", handler(controller.FindAssetsByCoinList))
}

// baseServer
// TODO 基础公共接口模块
func baseServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(verify.VerifyJWT())
		// 获取商户角色
		r.GET("/base/merchant/roles", handler(controller.GetBaseMerchantRoles))
		r.POST("/uploadFile", handler(controller.UpLoadFiles))
		// 第一次密码设置
		r.POST("/base/reset/firstPassword", handler(controller.NewPassword))
		// 检查是否为新账号
		r.POST("/base/check/isNewAccount", handler(controller.CheckNewAccount))
		// 获取个人信息
		r.GET("/base/personal/info", handler(controller.GetUserPersonal))
		// 退出登录
		r.POST("/base/personal/logout", handler(controller.Logout))
		// 获取角色
		r.GET("/base/sys/roles", handler(controller.GetBaseSysRoles))
		// 根据角色获取菜单
		r.GET("/base/menus/rid", handler(controller.GetBaseMenuByRid))
		// 根据商户Id获取业务线
		r.GET("/base/service/byMid", handler(controller.FindMerchantHaveAllService))
		// 根据商户Id获取业务线
		r.GET("/base/allService", handler(controller.FindAllService))
		// 获取单位
		r.GET("/base/allUnit", handler(controller.FindAllUnit))
		// 获取单位
		r.GET("/base/allAudit", handler(controller.FindAllAudit))
		// 获取全部主链币
		r.GET("/base/allChain", handler(controller.FindAllChain))
		// 获取全部主链币
		r.GET("/base/allChain", handler(controller.FindAllChain))
		// 获取全部审核类型
		r.GET("/base/auditTypeList", handler(controller.FindAuditStateList))
		// 获取全部审核结果
		r.GET("/base/auditResultList", handler(controller.FindAuditResultList))
		// 获取全部订单交易类型
		r.GET("/base/billTxTypeList", handler(controller.FindBillTxTypeList))
		// 根据商户ID查询业务线
		r.GET("/base/midService", handler(controller.FindMidService))
		// 根据业务线ID查询链、币种
		r.GET("/base/sidCoinList", handler(controller.FindSidCoinList))
		// TODO 我的模块
		// 更新密码
		r.POST("/update/password", handler(controller.UpdateOurPassword))
		// 更新自己的信息
		r.POST("/update/ourInfo", handler(controller.UpdateOurInfo))
		// 获取我的
		r.GET("/my/getMenu", handler(controller.GetMenuByRole))
	}
}

// 钱包 回调接口
func blockChainCallBackServer(e *echo.Echo) {
	r := e.Group("/custody")
	{
		// 参数 验签
		//r.Use(verify.CheckApiParamSign())
		r.POST("/blockchain/callback", handler(controller.BlockchainCallback))             // 提现回调地址
		r.POST("/blockchain/incomecallback", handler(controller.BlockchainIncomeCallback)) // 充值回调地址

	}
}

// loginServer
// TODO 登录模块
func loginServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 获取手机区域列表
		r.GET("/phone/code", handler(controller.FindPhoneCodeAll))
		// 账号密码登录
		r.POST("/login/password", handler(controller.LoginByPassword))
		// 验证码登录
		r.POST("/login/code", handler(controller.LoginByCode))
		// 检查账号是否有效
		r.POST("/check/account", handler(controller.VerifyDataAccount))
		// 忘记密码，重置
		r.POST("/reset/password", handler(controller.ResetPassword))
		// 发送登录验证码
		r.POST("/send/loginCode", handler(controller.SendLoginCode))
		// 发送重置密码验证码
		r.POST("/send/reset/pwdCode", handler(controller.SendResetPwd))
	}
}

// personnelServer
// TODO 权限-人员管理
func personnelServer(e *echo.Echo) {

	// 人员管理
	p := e.Group("/admin")
	{
		p.Use(mw.JWTWithConfig(GetJWTConfig()))
		p.Use(cd.CasbinHandler())
		p.Use(verify.VerifyJWT())
		p.Use(verify.VerifyCDN())
		// 中间件校验账号
		p.Use(verify.VerifyAccount())
		// 中间件校验账
		// 人员配置
		p.GET("/personnel/list", handler(controller.GetUserList))
		// 根据个人获取人员
		p.GET("/personnel/getUserById", handler(controller.GetUserById))
		// 新增人员
		p.POST("/personnel/addUser", handler(controller.AddAdminUserInfo))
		// 更新人员
		p.POST("/personnel/updateUser", handler(controller.UpdateAdminUserInfo))
		// 更新账号状态
		p.POST("/personnel/updateState", handler(controller.UpdateAdminUserState))
		// 修改添加超级审核员
		p.POST("/personnel/saveSuperAudit", handler(controller.SaveSuperAudit))
		// 根据用户Id获取业务线
		p.GET("/personnel/findService", handler(controller.GetSuperAudit))
		// 删除人员
		p.POST("/personnel/delete", handler(controller.DelAdminUserInfo))
	}
}

// merchantSubServer
// TODO 商户子账号管理
func merchantSubServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询子账号
		r.GET("/merchantSub/list", handler(controller.GetSubUserList))
		// 根据个人获取人员
		r.GET("/merchantSub/getUserById", handler(controller.GetSubUserInfo))
		// 更新人员
		r.POST("/merchantSub/updateUser", handler(controller.UpdateSubUserInfo))
		// 更新账号状态
		r.POST("/merchantSub/updateState", handler(controller.FreezeOrThawSubUserInfo))
		// 删除人员
		r.POST("/merchantSub/delete", handler(controller.DelSubUserInfo))
		// 一键清除账号异常
		r.POST("/merchantSub/clearAllErr", handler(controller.ClearSubInfoErr))
		// 清除账号异常
		r.POST("/merchantSub/clearErrById", handler(controller.ClearSubInfoErrById))
		// 用户操作记录
		r.GET("/merchantSub/operate", handler(controller.FindUserOperateUId))
	}
}

// merchantMainServer
// TODO 商户主账号管理
func merchantMainServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询主账号
		r.GET("/merchantMain/list", handler(controller.GetMainUserList))
		// 通过Id查看查看主账号详情
		r.GET("/merchantMain/getMainUserById", handler(controller.GetMainUserById))
		// 查看链情况
		r.GET("/merchantMain/chains", handler(controller.GetChainsInfo))
		// 查看人员角色详情
		r.GET("/merchantMain/roleInfo", handler(controller.GetRoleInfo))
		// 清除账号异常
		r.POST("/merchantMain/clearErrById", handler(controller.ClearSubInfoErrById))
	}
}

// assetsServer
// TODO 商户资产管理
func assetsServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		////r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询商户资产
		r.GET("/assets/list", handler(controller.GetAssetsList))
		// 查询商户资产折线图
		r.GET("/assets/line", handler(controller.GetAssetsLine))
		// 查询商户资产饼图
		r.GET("/assets/ring", handler(controller.GetAssetsRing))
	}
}

// merchantSubServer
// TODO 商户链路管理
func merchantChainsServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询商户链路列表
		r.GET("/merchantChains/list", handler(controller.GetMerchantChainList))
		// 根据Id商户链路列表
		r.GET("/merchantChains/getInfo", handler(controller.GetMerchantChainInfo))
		// 新增商户链路
		r.POST("/merchantChains/save", handler(controller.SaveMerchantChainInfo))
		// 更新商户链路
		r.POST("/merchantChains/updateInfo", handler(controller.UpdateMerchantChainInfo))
		// 更新商户链路状态
		r.POST("/merchantChains/updateState", handler(controller.FreezeOrThawMerchantChainInfo))
		// 删除商户链路
		r.POST("/merchantChains/delete", handler(controller.DelMerchantChainInfo))
	}
}

//套餐接口
func packageServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		r.POST("/package/item", handler(controller.CreateNewPackage))              //增加套餐
		r.DELETE("/package/item", handler(controller.DeletePackageItem))           //删除套餐
		r.POST("/package/update", handler(controller.UpdatePackageItem))           //修改套餐
		r.GET("/package/list", handler(controller.SearchPackageList))              //查询套餐列表
		r.GET("/package/item", handler(controller.SearchPackageItemInfo))          //查询Id套餐详情
		r.GET("/package/screen/list", handler(controller.SearchPackageScreenList)) //筛选列表
		//r.GET("/mch/package/item", financeHandler(controller.SearchMchPackageItemInfo))   //商户 查询套餐详情
	}
	rMCH := e.Group("/admin")
	{
		rMCH.GET("/mch/package/item", handler(controller.SearchMchPackageItemInfo)) //商户 查询套餐详情
	}
}

//业务线接口
func businessServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		r.POST("/business/item", handler(controller.CreateNewBusiness))      //增加业务线
		r.DELETE("/business/item", handler(controller.DeleteBusinessItem))   //删除业务线
		r.POST("/business/update", handler(controller.UpdateBusinessItem))   //修改业务线
		r.GET("/business/list", handler(controller.SearchBusinessList))      //查询业务线列表
		r.GET("/business/item", handler(controller.SearchBusinessItemInfo))  //查询业务线套餐详情
		r.POST("/business/operate", handler(controller.ActionBusinessItem))  //操作业务线，冻结/解冻
		r.GET("/business/pinfo", handler(controller.SearchBusinessItemInfo)) //套餐费用详情接口
		r.GET("/business/sinfo", handler(controller.BusinessSecurity))       //安全信息
		r.GET("/business/logs", handler(controller.BusinessOperateLogList))  //操作日志列表
		r.POST("/business/cs", handler(controller.ResetClientIdAndSecret))   //重置密钥CLIENT_ID/SECRET

	}
}

//币接口
func coinServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		r.GET("/coin/list", handler(controller.SearchCoinList))       //主链币列表
		r.GET("/subcoin/list", handler(controller.SearchSubCoinList)) //代币列表
	}
}

//商户接口
func merchantServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 中间件校验账
		r.GET("/merchant/apply/list", handler(controller.SearchApplyList)) //商户申请列表
		//r.POST("/merchant/image/update", financeHandler(controller.UpdateMerchantImage)) //商户列表
		r.POST("/merchant/operate", handler(controller.ActionMerchantItem)) //操作商户申请，通过/拒绝
		r.GET("/merchant/list", handler(controller.SearchMerchantList))     //商户列表
		r.GET("/apply/image", handler(controller.GetApplyImageInfo))        //获取认证图片/合同详情
		r.GET("/merchant/image", handler(controller.GetMerchantImageInfo))  //获取认证图片/合同详情
		r.GET("/merchant/item", handler(controller.SearchMerchantInfo))     //查询编辑详情
		r.POST("/merchant/item", handler(controller.UpdateMerchantItem))    //编辑商户
		r.POST("/merchant/push/one", handler(controller.PushMerchantItem))  //推送财务审核
		r.POST("/merchant/push/all", handler(controller.PushMerchantAll))   //一键推送财务审核
	}
}

//财务管理
func financeServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		r.GET("/finance/check/list", handler(controller.SearchFinanceList)) //财务审核列表
		r.POST("/finance/operate", handler(controller.ActionFinanceItem))   //操作财务审核申请，解冻冻结资产/解冻冻结
		r.GET("/finance/img", handler(controller.FinanceItemImg))           //获取商户图片（认证图片/合同图片/时间）
		r.POST("/finance/item", handler(controller.UpdateFinanceItem))      //编辑商户（认证图片/合同图片/时间）
		r.GET("/finance/logs", handler(controller.FinanceLockLogList))      //冻结日志详情
	}
}

//商户账单-业务线订单
func businessOrderServer(e *echo.Echo) {
	r0 := e.Group("/admin")
	{
		r0.POST("/business/order/item", handler(controller.AccountBusinessRenew))                //业务线续费（商户发起）
		r0.GET("/business/combo/list", handler(controller.SearchBusinessOrderList))              //订单列表
		r0.POST("/business/order/account/operate", handler(controller.ActionOrderItemByAccount)) //业务线订单，同意/拒绝（商户操作）
	}
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		//r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		r.GET("/business/order/list", handler(controller.SearchBusinessOrderList))          //订单列表
		r.GET("/business/order/down", handler(controller.DownBusinessOrderList))            //订单列表导出
		r.POST("/business/order/admin/operate", handler(controller.ActionOrderItemByAdmin)) //业务线订单，通过/拒绝（管理后台操作）
	}
}

// incomeServer
// TODO 收益户管理
func incomeServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		// 中间件配置
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		//财务收益户
		r.GET("/income/list", handler(controller.GetIncomeList))
		//财务收益户
		r.GET("/income/chart", handler(controller.IncomeChartInfo))
		//收益户导出
		r.POST("/income/export", handler(controller.ExportIncomeList))
	}
}

// auditServer
// TODO 审核管模块
func auditServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询订单
		r.GET("/orders/list", handler(controller.FindOrderList))
		// 查询各个类型订单数量
		//r.POST("/orders/status", financeHandler(controller.CountOrderStatus))
		// 审核进度查询
		//r.POST("/orders/plan/detail", financeHandler(controller.FinOrderPlanDetail))
		// 订单通过
		r.POST("/orders/update/pass", handler(controller.UpdatePassOrderInfo))
		// 订单解冻
		r.POST("/orders/update/thaw", handler(controller.UpdateThawOrderInfo))
		// 订单冻结
		r.POST("/orders/update/freeze", handler(controller.UpdateFreezeOrderInfo))
		// 订单拒绝
		r.POST("/orders/update/refuse", handler(controller.UpdateRefuseOrderInfo))
		// 一键通过
		r.POST("/orders/update/all", handler(controller.UpdateAllOrder))
		// 待审核订单导出
		r.POST("/orders/export", handler(controller.FindOrderExport))
	}
}

// billServer
// TODO 账单管理
func billServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		// 查询账单
		r.GET("/bill/list", handler(controller.FindBillList))
		// 导出账单
		r.GET("/bill/export", handler(controller.BillExcelExport))
		// 账单各类金额
		r.GET("/bill/state/balance", handler(controller.FindBillBalance))
		// 账单重推
		r.GET("/bill/push", handler(controller.PushBill))
	}
}

// chainBillServer
// TODO 链上订单管理
func chainBillServer(e *echo.Echo) {
	r := e.Group("/admin")
	{
		r.Use(mw.JWTWithConfig(GetJWTConfig()))
		r.Use(cd.CasbinHandler())
		r.Use(verify.VerifyJWT())
		r.Use(verify.VerifyCDN())
		// 中间件校验账号
		r.Use(verify.VerifyAccount())
		//链上订单查询
		r.GET("/chain/bill/list", handler(controller.FindChainBillList))
		//链上订单导出
		r.POST("/chain/bill/export", handler(controller.FindChainBillExport))
		//链上订单回滚
		r.POST("/chain/bill/reback", handler(controller.FindChainBillReBack))
		//链上订单重推
		r.POST("/chain/bill/repush", handler(controller.FindChainBillRePush))
	}
}

/**
 * 自定义Context的Handler
 */
func handler(h HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*handler2.Context)
		return h(ctx)
	}
}

// GetJWTConfig
// 中间件配置自定义类型JWT
// 获取JWT配置
func GetJWTConfig() mw.JWTConfig {
	// 中间件配置自定义类型JWT
	return mw.JWTConfig{
		Claims: &domain.JwtCustomClaims{},
		ErrorHandlerWithContext: func(err error, context echo.Context) error {
			context.JSON(http.StatusOK, map[string]interface{}{
				"code": 401,
				"msg":  global.MsgWarnJwtErr,
			})
			return errors.New(global.MsgWarnJwtErr)
		},
		SigningKey: []byte(auth.PrivateKey),
	}
}
