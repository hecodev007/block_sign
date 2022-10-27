package po

import "bosDataServer/common/db"

// 地址信息
type UserInfo struct {
	Id           int64  `json:"id,omitempty" gorm:"column:id"`
	TrxNotifyUrl string `json:"trx_notify_url,omitempty" gorm:"column:trx_notify_url"`
}

func (u *UserInfo) TableName() string {
	return "users"
}

//type Users struct {
//	Id           int       `xorm:"not null pk autoincr INT(10)"`
//	CreatedAt    time.Time `xorm:"TIMESTAMP"`
//	UpdatedAt    time.Time `xorm:"TIMESTAMP"`
//	DeletedAt    time.Time `xorm:"index TIMESTAMP"`
//	MerchantName string    `xorm:"unique VARCHAR(128)"`
//	TrxNotifyUrl string    `xorm:"default '' VARCHAR(255)"`
//	Description  string    `xorm:"default '' VARCHAR(255)"`
//}
type AddressesInfo struct {
	Id       int64  `json:"id,omitempty" gorm:"column:id"`
	Address  string `json:"address,omitempty" gorm:"column:address"`
	UserId   int64  `json:"user_id,omitempty" gorm:"column:user_id"`
	CoinType string `json:"coin_type,omitempty" gorm:"column:coin_type"`
	Status   string `json:"status,omitempty" gorm:"column:status"`
}

func (u *AddressesInfo) TableName() string {
	return "addresses"
}

func FindUserInfos() (list []*UserInfo, err error) {
	err = db.UserConn.Find(&list)
	return
}

func FindAddressesInfos(cointype string) (list []*AddressesInfo, err error) {
	//db.UserConn.ShowSQL(true)
	err = db.UserConn.Where("coin_type = ? and status = 'used'", cointype).Find(&list)
	return
}

// 合约信息
type ContractInfo struct {
	Id              int64  `json:"id,omitempty" gorm:"column:id"`
	Name            string `json:"name,omitempty" gorm:"column:name"`                         // 合约名称
	ContractAddress string `json:"contract_address,omitempty" gorm:"column:contract_address"` // 合约地址
	Decimal         int    `json:"decimal,omitempty" gorm:"column:decimal"`                   // 精度
	CoinType        string `json:"coin_type,omitempty" gorm:"column:coin_type"`               // 币种名称
	Invaild         int    `json:"invaild,omitempty" gorm:"column:invaild"`                   // 0 有效 1 无效
}

func (o *ContractInfo) TableName() string {
	return "contract_info"
}

// 删除区块
func DeleteContractInfo(contractAddr string) (err error) {
	_, err = db.UserConn.Where("contract_address=?", contractAddr).Delete(new(ContractInfo))
	return
}

func FindContractInfos(coinType string) (list []*ContractInfo, err error) {

	list = make([]*ContractInfo, 0)
	err = db.UserConn.Where("coin_type = ? and invaild = ?", coinType, 0).Find(&list)
	l := &ContractInfo{
		Id:0,
		Name :"bos",
		ContractAddress:"eosio.token",
		Decimal:4,
		CoinType:"bos",
		Invaild:0,
	}
	list = append(list,l)
	return
}

// hash 获取交易数据
func SelectNameByAddress(contractAddr string) (string, error) {
	c := new(ContractInfo)
	_, err := db.UserConn.Where("contract_address = ?", contractAddr).Get(c)
	return c.Name, err
}

// 插入块数据
// return 影响行
func InsertContractInfo(c *ContractInfo) (int64, error) {
	_, err := db.UserConn.Insert(c)
	return c.Id, err
}
