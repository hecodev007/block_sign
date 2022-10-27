package po

import "avaxDataServer/db"

// 地址信息
type UserInfo struct {
	ID        int64  `json:"id,omitempty" gorm:"column:id"`
	NotifyUrl string `json:"trx_notify_url,omitempty" gorm:"column:trx_notify_url"`
}

type AddressesInfo struct {
	ID       int64  `json:"id,omitempty" gorm:"column:id"`
	Address  string `json:"address,omitempty" gorm:"column:address"`
	UserID   int64  `json:"user_id,omitempty" gorm:"column:user_id"`
	CoinType string `json:"coin_type,omitempty" gorm:"column:coin_type"`
	Status   string `json:"status,omitempty" gorm:"column:status"`
}

func (u *UserInfo) TableName() string {
	return "users"
}

func (u *AddressesInfo) TableName() string {
	return "addresses"
}

func FindUserInfos() ([]UserInfo, error) {
	var us []UserInfo
	err := db.UserDB.DB.Select([]string{"id", "trx_notify_url"}).Find(&us).Error
	if err != nil {
		return nil, err
	}

	return us, nil
}

func FindAddressesInfos(cointype string) ([]AddressesInfo, error) {
	var as []AddressesInfo
	err := db.UserDB.DB.Where("coin_type = ? and status = 'used'", cointype).Find(&as).Error
	if err != nil {
		return nil, err
	}

	return as, nil
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
func DeleteContractInfo(contractAddr string) error {

	err := db.UserDB.DB.Exec("delete from contract_info where contract_address = ?", contractAddr).Error
	if err != nil {
		return err
	}

	return nil
}

func FindContractInfos(coinType string) ([]ContractInfo, error) {
	var cs []ContractInfo
	err := db.UserDB.DB.Where("coin_type = ? and invaild = ?", coinType, 0).Find(&cs).Error
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// hash 获取交易数据
func SelectNameByAddress(contractAddr string) (string, error) {
	//o := orm.NewOrm()
	//var maps []orm.Params
	//nums, err := o.Raw("select name from contract_info where contract_address = ?", contractAddr).Values(&maps)
	//if err == nil && nums > 0 {
	//	return maps[0]["name"].(string), nil
	//}
	//
	//return "eth", fmt.Errorf("don't find token name ")
	c := &ContractInfo{}
	if err := db.UserDB.DB.Select([]string{"name"}).Where("contract_address = ?", contractAddr).First(c).Error; err != nil {
		return "eth", err
	}

	return c.Name, nil
}

// 插入块数据
// return 影响行
func InsertContractInfo(c *ContractInfo) (int64, error) {
	//o := orm.NewOrm()
	//res, err := o.Raw("insert into contract_info(name,contract_address,deciaml) values(?,?,?)",
	//	c.Name, c.ContractAddress, c.Decimal).Exec()
	//if err == nil {
	//	num, _ := res.RowsAffected()
	//	return num, nil
	//}
	//return 0, nil
	if err := db.UserDB.DB.Create(c).Error; err != nil {
		return 0, err
	}
	return c.Id, nil
}
