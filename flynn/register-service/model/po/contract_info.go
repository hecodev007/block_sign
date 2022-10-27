package po

import "github.com/group-coldwallet/flynn/register-service/db"

// 合约信息
type ContractInfo struct {
	Id              int64  `json:"id,omitempty" gorm:"column:id"`
	Name            string `json:"name,omitempty" gorm:"column:name"`                         // 合约名称
	ContractAddress string `json:"contract_address,omitempty" gorm:"column:contract_address"` // 合约地址
	Decimal         int    `json:"decimal,omitempty" gorm:"column:decimal"`                   // 精度
	CoinType        string `json:"coin_type,omitempty" gorm:"column:coin_type"`               // 币种名称
	Invaild         int    `json:"invaild,omitempty" gorm:"column:invaild"`                   // 0 有效 1 无效
}

// 插入块数据
// return 影响行
func InsertContractInfo(c *ContractInfo) error {
	_, err := db.UserConn.InsertOne(c)
	if err != nil {
		return err
	}
	return nil
}

func FindContract(coinName, mainCoinName, contractAddress string) ([]ContractInfo, error) {
	var addr []ContractInfo
	err := db.UserConn.Where("name = ? and coin_type = ? and contract_address = ?", coinName, mainCoinName, contractAddress).Find(&addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

/*
更新为有效
*/
func UpdateContractInvaild(coinName, mainCoinName, contractAddress string, invaild int) error {
	var a ContractInfo
	a.Invaild = invaild
	_, err := db.UserConn.Cols("invaild").
		Where("name = ? and coin_type = ? and contract_address = ?", coinName, mainCoinName, contractAddress).
		Update(&a)
	if err != nil {
		return err
	}
	return nil
}

func DeleteContractInfo(coinName, mainCoinName, contractAddress string, invaild int) error {
	result := new(ContractInfo)

	_, err := db.UserConn.Where("name = ? and coin_type = ? and contract_address = ? and invaild = ? ",
		coinName, mainCoinName, contractAddress, invaild).Delete(result)
	if err != nil {
		return err
	}
	return nil
}
