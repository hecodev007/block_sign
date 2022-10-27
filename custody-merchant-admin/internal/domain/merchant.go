package domain

type ApplyReqInfo struct {
	Id           int64  `json:"id" query:"id"`
	ContactStr   string `json:"contact_str" query:"contact_str"` //联系方式
	AccountId    string `json:"account_id" query:"account_id"`
	AccountName  string `json:"account_name" query:"account_name"`
	CardNum      string `json:"card_num" query:"card_num"` //证件号
	VerifyStatus string `json:"verify_status" query:"verify_status"`
	VerifyResult string `json:"verify_result" query:"verify_result"`
	Limit        int    `json:"limit"  query:"limit" description:"查询条数" example:"10"`
	Offset       int    `json:"offset" query:"offset" description:"查询起始位置" example:"0"`
}

type MerchantReqInfo struct {
	Id             int64  `json:"id" query:"id"`
	ContactStr     string `json:"contact_str" query:"contact_str"` //联系方式
	AccountId      string `json:"account_id" query:"account_id"`
	AccountName    string `json:"account_name" query:"account_name"`
	CardNum        string `json:"card_num" query:"card_num"`                 //证件号
	RealNameStatus string `json:"real_name_status" query:"real_name_status"` //实名状态 hadreal-已经实名，unreal-未实名
	FvStatus       string `json:"fv_status" query:"fv_status"`               //财务审核 （agree-通过，refuse-拒绝，wait-未处理）
	LockStatus     string `json:"lock_status" query:"lock_status"`           //冻结状态（unlock-正常，lock-异常）
	RealNameStart  string `json:"real_name_start" query:"real_name_start"`   //实名开始时间
	RealNameEnd    string `json:"real_name_end" query:"real_name_end"`       //实名结束时间
	Limit          int    `json:"limit"  query:"limit" description:"查询条数" example:"10"`
	Offset         int    `json:"offset" query:"offset" description:"查询起始位置" example:"0"`
}

type MerchantImageInfo struct {
	Id int64 `json:"id" query:"id"`
	//ImageType string `json:"image_type" query:"image_type"` //identity-身份，business-营业执照，contract-合同
}

type MerchantOperateInfo struct {
	Id        int64  `json:"id" query:"id"`
	Operate   string `json:"operate" query:"operate"`       //agree-同意，refuse-拒绝
	TestEnd   string `json:"test_end" query:"test_end"`     //测试截止时间
	Remark    string `json:"remark" query:"remark"`         //审核备注
	StartTime string `json:"start_time" query:"start_time"` //
	EndTime   string `json:"end_time" query:"end_time"`     //
}

type MerchantImgInfo struct {
	Id              int64    `json:"id" query:"id"`
	IdCardPicture   []string `json:"id_card_picture" query:"id_card_picture"`   //agree-同意，refuse-拒绝
	BusinessPicture []string `json:"business_picture" query:"business_picture"` //测试截止时间
	ContractPicture []string `json:"contract_picture" query:"contract_picture"` //审核备注

}

type MerchantUpdateInfo struct {
	Id int64 `json:"id" form:"id"`
	//ApplyId         int64  `json:"apply_id" form:"apply_id"`
	//MerchantId      int64  `json:"merchant_id" form:"merchant_id"`
	Name            string `json:"name" form:"name"`
	Phone           string `json:"phone" form:"phone"`
	PhoneCode       string `json:"phone_code" form:"phone_code"`
	Email           string `json:"email" form:"email"`
	IdCardNum       string `json:"id_card_num" form:"id_card_num"`
	PassportNum     string `json:"passport_num" form:"passport_num"`
	Sex             int    `json:"sex" form:"sex"`
	ContractStartAt string `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" form:"contract_end_at"`
	TestEnd         string `json:"test_end" gorm:"test_end"`
	Remark          string `json:"remark" gorm:"remark"`
}

type MerchantInfo struct {
	Id              int64  `json:"id" form:"id"`
	AccountId       int64  `json:"account_id" form:"account_id"`
	AccountName     string `json:"account_name" form:"account_name"`
	Phone           string `json:"phone" form:"phone"`
	Email           string `json:"email" form:"email"`
	IdCardNum       string `json:"id_card_num" form:"id_card_num"`
	PassportNum     string `json:"passport_num" form:"passport_num"`
	CreatedAt       string `json:"created_at" form:"created_at"`
	CardType        string `json:"card_type" form:"card_type"`
	CoinName        string `json:"coin_name" form:"coin_name"`
	TradeType       string `json:"trade_type,omitempty" gorm:"column:trade_type"` //交易类型(open-开通，2renew-二次续费，3renew-三次续费。。。。)
	VerifyStatus    string `json:"verify_status" form:"verify_status"`
	VerifyAt        string `json:"verify_at" form:"verify_at"`
	VerifyUser      string `json:"verify_user" form:"verify_user"`
	VerifyResult    string `json:"verify_result" form:"verify_result"`
	AccountStatus   string `json:"account_status" form:"account_status"` //用户状态（formal-正式，test-测试）
	TestEnd         string `json:"test_end" form:"test_end"`
	FvStatus        string `json:"fv_status" form:"fv_status"`     //财务审核状态
	FVRemark        string `json:"fv_remark" form:"fv_remark"`     //财务操作备注
	LockStatus      int    `json:"lock_status" form:"lock_status"` //账户冻结状态
	LockRemark      string `json:"lock_remark" form:"lock_remark"` //账户冻结备注
	RealNameStatus  int    `json:"real_name_status" form:"real_name_status"`
	RealNameAt      string `json:"real_name_at" form:"real_name_at"`
	ContractStartAt string `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" form:"contract_end_at"`
	IsPush          string `json:"is_push" form:"is_push"`
	PushAble        string `json:"push_able" form:"push_able"`
}

type FinanceListInfo struct {
	Id              int64  `json:"id" form:"id"`
	AccountId       int64  `json:"account_id" form:"account_id"`
	AccountName     string `json:"account_name" form:"account_name"`
	Phone           string `json:"phone" form:"phone"`
	Email           string `json:"email" form:"email"`
	IdCardNum       string `json:"id_card_num" form:"id_card_num"`
	PassportNum     string `json:"passport_num" form:"passport_num"`
	CreatedAt       string `json:"created_at" form:"created_at"`
	RealNameStatus  string `json:"real_name_status" form:"real_name_status"`
	RealNameAt      string `json:"real_name_at" form:"real_name_at"`
	ContractStartAt string `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" form:"contract_end_at"`
	FvStatus        string `json:"fv_status" form:"fv_status"`           //财务审核状态
	FVRemark        string `json:"fv_remark" form:"fv_remark"`           //财务操作备注
	LockStatus      string `json:"lock_status" form:"lock_status"`       //账户冻结状态
	LockRemark      string `json:"lock_remark" form:"lock_remark"`       //账户冻结备注
	AccountStatus   int    `json:"account_status" form:"account_status"` //用户状态（0-正式，1-测试）
	TestEnd         string `json:"test_end" gorm:"test_end"`
	TradeType       string `json:"trade_type" gorm:"column:trade_type"`    //交易类型(open-开通，2renew-二次续费，3renew-三次续费。。。。)
	IsLockFinance   int    `json:"is_lock_finance" form:"is_lock_finance"` //用户状态（0-正式，1-测试）
	IsLock          int    `json:"is_lock" form:"is_lock"`                 //用户状态（0-正式，1-测试）
}

type MerchantListInfo struct {
	SerialNo        int    `json:"serial_no" gorm:"serial_no"`
	Id              int64  `json:"id" gorm:"id"`
	AccountId       int64  `json:"account_id" gorm:"account_id"`
	IsTest          int    `json:"is_test" gorm:"is_test"`
	IsPush          int    `json:"is_push" gorm:"is_push"`
	Name            string `json:"name" gorm:"name"`
	Phone           string `json:"phone" gorm:"phone"`
	Email           string `json:"email" gorm:"email"`
	IdCardNum       string `json:"id_card_num" gorm:"id_card_num"`
	PassportNum     string `json:"passport_num" gorm:"passport_num"`
	RealNameStatus  int    `json:"real_name_status" gorm:"real_name_status"` //实名状态 0-未实名，1-已实名
	RealNameAt      string `json:"real_name_at" gorm:"real_name_at"`
	TestEnd         string `json:"test_end" gorm:"test_end"`
	CreatedAt       string `json:"created_at" gorm:"created_at"`
	FvStatus        string `json:"fv_status" gorm:"fv_status"` // 财务审核 wait-未审核，agree-已审核通过，refuse-已拒绝
	ContractStartAt string `json:"contract_start_at" gorm:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" gorm:"contract_end_at"`
}

type ApplyInfo struct {
	Id              int64  `json:"id" form:"id"`
	AccountId       int64  `json:"account_id" form:"account_id"`
	Name            string `json:"name" form:"name"`
	Phone           string `json:"phone" form:"phone"`
	Email           string `json:"email" form:"email"`
	IdCardNum       string `json:"id_card_num" form:"id_card_num"`
	PassportNum     string `json:"passport_num" form:"passport_num"`
	CreatedAt       string `json:"created_at" form:"created_at"`
	CardType        string `json:"card_type" form:"card_type"`
	CoinName        string `json:"coin_name" form:"coin_name"`
	VerifyStatus    string `json:"verify_status" form:"verify_status"`
	VerifyAt        string `json:"verify_at" form:"verify_at"`
	VerifyUser      string `json:"verify_user" form:"verify_user"`
	VerifyResult    string `json:"verify_result" form:"verify_result"`
	AccountStatus   string `json:"account_status" form:"account_status"` //用户状态（formal-正式，test-测试）
	TestEnd         string `json:"test_end" form:"test_end"`
	ContractStartAt string `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" form:"contract_end_at"`
	Remark          string `json:"remark" form:"remark"`
}
