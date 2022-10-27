package auditRole

type AuditRole struct {
	Id    int    `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Name  string `json:"name" gorm:"column:name"`
	State string `json:"remark" gorm:"column:remark"`
}

func (ar *AuditRole) TableName() string {
	return "audit_role"
}
