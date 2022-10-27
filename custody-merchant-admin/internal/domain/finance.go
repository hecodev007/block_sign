package domain

//ApplyUpImageReqInfo 财务审核编辑身份/合同图片
type ApplyImageReqInfo struct {
	Id              int64  `json:"id" query:"id"`
	IdCardPicture   string `json:"identity" form:"identity"` //身份证件
	BusinessPicture string `json:"business" form:"business"` //营业执照证
	ContractPicture string `json:"contract" form:"contract"` //合同图片
	ContractStartAt string `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string `json:"contract_end_at" form:"contract_end_at"`
}

//ApplyUpImageReqInfo 财务审核编辑身份/合同图片
type ApplyUpdateImageReqInfo struct {
	Id              int64    `json:"id" form:"id"`
	IdCardPicture   []string `json:"identity" form:"identity"` //身份证件
	BusinessPicture []string `json:"business" form:"business"` //营业执照证
	ContractPicture []string `json:"contract" form:"contract"` //合同图片
	ContractStartAt string   `json:"contract_start_at" form:"contract_start_at"`
	ContractEndAt   string   `json:"contract_end_at" form:"contract_end_at"`
}

//FinanceOperateInfo 财务审核冻结
type FinanceOperateInfo struct {
	Id          int64  `json:"id" form:"id"`
	Operate     string `json:"operate" form:"operate"`           //冻结用户和资产lock_user，解冻用户和资产unlock_user，冻结资产lock_asset，，解冻资产unlock_asset
	OperateId   string `json:"operate_id" form:"operate_id"`     //审核人id
	OperateName string `json:"operate_name" form:"operate_name"` // 审核人
	Remark      string `json:"remark" form:"remark"`             //审核备注
}
