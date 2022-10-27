package po

import "github.com/group-coldwallet/flynn/register-service/db"

type Addresses struct {
	Id       int64  `json:"id,omitempty" gorm:"column:id"`
	Address  string `json:"address,omitempty" gorm:"column:address"`
	UserId   int64  `json:"user_id,omitempty" gorm:"column:user_id"`
	CoinType string `json:"coin_type,omitempty" gorm:"column:coin_type"`
	Status   string `json:"status,omitempty" gorm:"column:status"`
}

func InsertWatchAddress(u []*Addresses) error {

	_, err := db.UserConn.Insert(u)
	if err != nil {
		return err
	}
	return nil

}
func FindAddresses(coinName string, userId int64, addresses []string) ([]Addresses, error) {
	var addr []Addresses
	err := db.UserConn.Where("user_id = ? and coin_type = ? ", userId, coinName).
		In("address", addresses).Find(&addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func UpdateAddressesStatus(coinName string, userId int64, addresses []string, status string) error {
	var addr Addresses
	addr.Status = status
	_, err := db.UserConn.Cols("status").
		Where("user_id = ? and coin_type = ? ", userId, coinName).
		In("address", addresses).
		Update(&addr)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAddresses(coinName string, userId int64, addresses []string, status string) error {
	result := new(Addresses)
	_, err := db.UserConn.Where("user_id = ? and coin_type = ? and status= ? ",
		userId, coinName, status).In("address", addresses).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
