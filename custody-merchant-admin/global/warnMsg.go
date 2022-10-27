package global

import "fmt"

const (
	MsgWarnParamsErr      = "传参错误"
	MsgWarnSqlInject      = "检测到SQL注入"
	MsgWarnSqlUpdate      = "更新失败"
	MsgWarnSysBuss        = "后台服务频繁,未处理,稍后请重试"
	MsgWarnSysParamRepeat = "新增数据重复"
	MsgWarnSysNotUserInfo = "没有用户数据"

	MsgWarnServiceNoPermission = "您对该业务线没有操作权限"
	MsgWarnNoRate              = "无折算汇率"
	MsgWarnNoCombo             = "不存在该套餐"
	MsgWarnNoService           = "业务线不存在"
	MsgWarnNoPreLevel          = "上一个层级不存在，无法跨层级添加"
	MsgWarnHaveLevel           = "该层级已经存在"
	MsgWarnHaveAssets          = "已经存在该资产"
	MsgWarnNoAssets            = "不存在该资产"
	MsgWarnAssetsLess          = "能使用的资产金额不足"

	MsgWarnNoBill = "没有这个账单"
	MsgWarnNoPush = "账单暂时无法重推"

	MsgWarnModelAdd          = "新增失败"
	MsgWarnModelDelete       = "删除失败"
	MsgWarnModelUpdate       = "更新失败"
	MsgWarnModelNil          = "暂无数据"
	MsgWarnModelErrorAccount = "账号或者密码错误"

	MsgWarnNoLine               = "已经用完，如果要继续使用，请尽快联系后台工作人员添加额度"
	MsgWarnNoAddr               = "已经用完，如果要继续使用，请尽快联系后台工作人员添加地址数"
	MsgWarnNoMonth              = "已经到期，如果要继续使用，请尽快联系后台工作人员续费"
	MsgWarnWillLine             = "将耗尽，如果要继续使用，请尽快联系后台工作人员添加额度"
	MsgWarnWillAddr             = "将耗尽，如果要继续使用，请尽快联系后台工作人员添加地址数"
	MsgWarnWillMonth            = "将到期，如果要继续使用，请尽快联系后台工作人员续费"
	MsgWithdrawLimitMinuteNums  = "5分钟内，提币金额过多"
	MsgWithdrawLimitHourNums    = "1小时内，提币金额过多"
	MsgWithdrawLimitMinuteCount = "5分钟内，提币次数超过了限制"
	MsgWithdrawLimitHourCount   = "1小时内，提币次数超过了限制"
	MsgWithdrawLimitClose       = "该业务线对外提币，已经关闭，无法提币"
	MsgWithdrawLimitWhite       = "对外提币失败：白名单未开放"
	MsgTransferLimitEachNums    = "这笔转出金额已经超过了限制"
	MsgTransferLimitDayNums     = "今天转出金额已经超过了限制"
	MsgTransferLimitWeekNums    = "本周转出金额已经超过了限制"
	MsgTransferLimitMonthNums   = "本月转出金额已经超过了限制"

	MsgWarnUserLevelNoPermission = "该用户不负责该层级"
	MsgWarnUserHaveServiceConfig = "该用户已存在配置"
	MsgWarnUserHaveSuperadmin    = "超级管理员只能有一个"

	MsgWarnAuditThanNums = "审核数量大于关联数"
	MsgWarnNoThisLevel   = "不存在该层级"
	MsgWarnAuditParamErr = "审核启用门槛的参数不正确"

	MsgWarnNotUser        = "这个用户不存在"
	MsgWarnEmailFormatErr = "邮箱账号错误"
	MsgWarnPhoneFormatErr = "手机号错误"
	MsgWarnAccountErr     = "账号无效"
	MsgWarnDecryptErr     = "解密失败"
	MsgWarnCodeErr        = "验证码错误"
	MsgWarnPasswordNil    = "密码为空"

	MsgWarnPhoneCodeErr = "该账号密码输入5次错误,请联系管理员"
	MsgWarnEmailCodeErr = "该邮箱验证码输入5次错误,请联系管理员"
	MsgWarnPwdCodeErr   = "该手机验证码输入5次错误,请联系管理员"

	MsgWarnNoHaveServiceLevel     = "数据错误，该业务线不存在审核等级"
	MsgWarnNoHaveService          = "数据错误，该业务线无效"
	MsgWarnCloseServiceWithdrawal = "该业务线已经关闭提币"
)

const (
	OperationWarn                        = "非法操作"
	OperationErr                         = "操作失败,"
	OperationIsOtherServiceAudit         = "修改人员 %d,是该业务线的其他等级审核员，请先取消他的等级"
	OperationIsNotVisitorErr             = "%s 该用户不是游客"
	OperationIsNotFinanceErr             = "%s 该用户不是财务"
	OperationIsNotAuditErr               = "您不是审核员"
	OperationUserNotSuperAudit           = "您不是该业务线的超级审核员,无权解冻"
	OperationIsNotServiceAuthErr         = "没有权限操作该业务线"
	OperationServiceAuthTypeThanUsersErr = "终审制数大于审核员数"
	OperationAddOrderErr                 = "订单新增失败"
	OperationUpdateThawOrderErr          = "订单更新失败，非冻结状态"
	OperationUpdateNormalOrderErr        = "订单更新失败，非正常状态"
	OperationDelUserHaveServiceErr       = "该用户涉及 %s 业务线，由于审核终审制>审核人员，删除该用户审核终审制将失灵，请先对该业务线审核终审制进行修改"
	OperationIsServiceLevelFreeze        = "您在该业务线的审核操作被冻结"
)

const (
	MsgWarnUpdateYourSelf  = "无法更新操作自己"
	MsgWarnDelSuper        = "无法删除超级管理员"
	MsgWarnUpdateSuper     = "无法更新超级管理员"
	MsgWarnAccountIsSubErr = "这个用户不存在或者不是您的子账号"
)
const (
	DataIsMore             = "后台服务频繁,未处理,30秒后请重试"
	DataBusinessComboIsNil = "业务线%d没有套餐"
)
const (
	DataNoHaveErr = "%s 为空"
)

const (
	DataWarnCreateUserErr        = "创建商户失败"
	DataWarnNoDataErr            = "数据不存在"
	DataWarnNoMerchantErr        = "商户不存在"
	DataWarnNoPackageErr         = "套餐不存在"
	DataWarnParamErr             = "参数错误"
	DataWarnAccountUnableErr     = "商户已失效"
	DataWarnDataUnableErr        = "数据已失效"
	DataWarnUpdateDataErr        = "数据更新失败"
	DataWarnHadToFinanceErr      = "不可重复推送"
	DataWarnBatchPushFinanceErr  = "批量推送出错"
	DataWarnNoImageErr           = "数据图片信息不完整"
	DataWarnNoContractErr        = "数据合同信息不完整"
	DataWarnNoPushUserErr        = "无可推送财务审核商户"
	DataWarnVerifyErr            = "数据未被审核通过"
	DataWarnHadVerifySusErr      = "数据已被财务审核不可修改"
	DataWarnHadLockErr           = "重复冻结"
	DataWarnNoLockErr            = "未冻结，无需解冻"
	DataWarnHadVerifyErr         = "数据已被审核，不可重复操作"
	DataWarnNoOperateErr         = "不支持的操作"
	DataWarnCreateComboErr       = "创建业务线失败"
	DataWarnUpdateComboErr       = "更新业务线失败"
	DataWarnUpdateLockErr        = "冻结失败"
	DataWarnUpdateUnLockErr      = "解冻失败"
	DataWarnUnlockAccountLockErr = "账户冻结中，无法解冻资产"
	DataWarnNoMchComboErr        = "业务线钱包cliend不存在"
	DataWarnCreateAddressErr     = "生成地址失败"
	DataWarnOrderDeductErr       = "订单扣款失败"
	DataWarnBalanceErr           = "账户余额不足"
	DataWarnNoBelongErr          = "业务线不属于当前商户"
	DataWarnNoAccountVerifyErr   = "商户未审核该订单，不可操作"
	DataWarnAccountRefuseErr     = "商户已拒绝订单，不可操作"
	DataWarnAccountPassErr       = "商户已通过订单，不可操作"
	DataWarnNoAccountAgreeErr    = "商户未同意该订单，不可操作"
	DataWarnNoFeeErr             = "已不满足满年优惠，请重新检查订单"

	DataWarnNoTransferTypeErr = "请选择更新交易类型"
)

func OperationErrorText(format string, a ...interface{}) error {
	return fmt.Errorf(OperationErr+format, a)
}

func OperationError(format string) error {
	return fmt.Errorf(OperationErr + format)
}
