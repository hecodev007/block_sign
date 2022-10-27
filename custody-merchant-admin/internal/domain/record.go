package domain

type RecordReqInfo struct {
	Id         int64 `json:"id" query:"id"`
	BusinessId int64 `json:"business_id" query:"business_id"`
	Limit      int   `json:"limit"   query:"limit" description:"查询条数" example:"10"`
	Offset     int   `json:"offset"  query:"offset" description:"查询起始位置" example:"0"`
}

type RecordInfo struct {
	Operate      string `json:"operate" form:"column:operate"`
	OperatorName string `json:"operator_name" form:"column:operator_name"`
	Remark       string `json:"remark" form:"column:remark"`
	CreatedAt    string `json:"created_at" form:"created_at"`
}

type FinanceRecordInfo struct {
	IsLock        int    `json:"is_lock"  form:"column:is_lock"`                 //是否冻结
	IsLockFinance int    `json:"is_lock_finance"  form:"column:is_lock_finance"` //是否冻结资产
	OperatorName  string `json:"operator_name" form:"column:operator_name"`
	Remark        string `json:"remark" form:"column:remark"`
	CreatedAt     string `json:"created_at" form:"created_at"`
}
