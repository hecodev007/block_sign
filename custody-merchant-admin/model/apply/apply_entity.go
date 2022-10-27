package apply

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

//Entity 商户申请列表
type Entity struct {
	Db              *orm.CacheDB `json:"-" gorm:"-"`
	Id              int64        `json:"id" gorm:"id"`
	AccountId       int64        `json:"account_id" gorm:"account_id"`
	Name            string       `json:"name" gorm:"name"`
	Phone           string       `json:"phone" gorm:"phone"`
	Email           string       `json:"email" gorm:"email"`
	CoinName        string       `json:"coin_name" gorm:"coin_name"`
	VerifyStatus    string       `json:"verify_status" gorm:"verify_status"` //商户申请审核状态 空-未审核，agree-已审核通过，refuse-已拒绝
	VerifyAt        *time.Time   `json:"verify_at" gorm:"verify_at"`
	VerifyUser      string       `json:"verify_user" gorm:"verify_user"`
	TestEnd         string       `json:"test_end" gorm:"test_end"`
	Introduce       string       `json:"introduce" gorm:"introduce"` //公司介绍
	ContractStartAt *time.Time   `json:"contract_start_at" gorm:"contract_start_at"`
	ContractEndAt   *time.Time   `json:"contract_end_at" gorm:"contract_end_at"`
	IdCardNum       string       `json:"id_card_num" gorm:"id_card_num"`
	PassportNum     string       `json:"passport_num" gorm:"passport_num"`
	IdCardPicture   string       `json:"id_card_picture" gorm:"id_card_picture"`   //身份证件
	BusinessPicture string       `json:"business_picture" gorm:"business_picture"` //营业执照证
	ContractPicture string       `json:"contract_picture" gorm:"contract_picture"` //合同图片
	RealNameStatus  int          `json:"real_name_status" gorm:"real_name_status"` //实名状态 0-未实名，1-已实名
	RealNameAt      *time.Time   `json:"real_name_at" gorm:"real_name_at"`
	Remark          string       `json:"remark" gorm:"remark"`
	CreatedAt       time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" gorm:"updated_at"`
}

func (e *Entity) TableName() string {
	return "apply_pending"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type ApplyDb struct {
	Db              *orm.CacheDB `json:"-" gorm:"-"`
	Id              int64        `json:"id" gorm:"id"`
	AccountId       int64        `json:"account_id" gorm:"account_id"`
	Name            string       `json:"name" gorm:"name"`
	Phone           string       `json:"phone" gorm:"phone"`
	Email           string       `json:"email" gorm:"email"`
	CoinName        string       `json:"coin_name" gorm:"coin_name"`
	VerifyStatus    string       `json:"verify_status" gorm:"verify_status"` //商户申请审核状态 空-未审核，agree-已审核通过，refuse-已拒绝
	VerifyAt        *time.Time   `json:"verify_at" gorm:"verify_at"`
	VerifyUser      string       `json:"verify_user" gorm:"verify_user"`
	TestEnd         string       `json:"test_end" gorm:"test_end"`
	Introduce       string       `json:"introduce" gorm:"introduce"` //公司介绍
	ContractStartAt *time.Time   `json:"contract_start_at" gorm:"contract_start_at"`
	ContractEndAt   *time.Time   `json:"contract_end_at" gorm:"contract_end_at"`
	IdCardNum       string       `json:"id_card_num" gorm:"id_card_num"`
	PassportNum     string       `json:"passport_num" gorm:"passport_num"`
	IdCardPicture   string       `json:"id_card_picture" gorm:"id_card_picture"`   //身份证件
	BusinessPicture string       `json:"business_picture" gorm:"business_picture"` //营业执照证
	ContractPicture string       `json:"contract_picture" gorm:"contract_picture"` //合同图片
	RealNameStatus  int          `json:"real_name_status" gorm:"real_name_status"` //实名状态 0-未实名，1-已实名
	RealNameAt      *time.Time   `json:"real_name_at" gorm:"real_name_at"`
	Remark          string       `json:"remark" gorm:"remark"`
	CreatedAt       time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" gorm:"updated_at"`
	IsTest          int          `json:"is_test" gorm:"is_test"`
}
