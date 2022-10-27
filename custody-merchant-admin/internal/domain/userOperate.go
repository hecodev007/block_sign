package domain

type OperateList struct {
	Id        int64  `json:"id" gorm:"column:id; PRIMARY_KEY"`
	UserId    int64  `json:"user_id" gorm:"column:user_id"`
	UserName  string `json:"user_name" gorm:"column:user_name"`
	CreatedBy string `json:"created_by" gorm:"column:created_by"`
	Content   string `json:"content" gorm:"column:content"`
	Platform  string `json:"platform" gorm:"column:platform"`
	CreatedAt string `json:"created_at" gorm:"created_at"`
}
