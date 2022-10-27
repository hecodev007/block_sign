package seek

import (
	"github.com/group-coldwallet/chaincore2/models"
	"strings"
)

// 关注地址列表, key: 合约地址
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {

}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	address = strings.ToLower(address)
	if WatchContractList[address] != nil {
		return
	}

	contract := new(models.ContractInfo)
	contract.Name = name
	contract.Decimal = decimal
	contract.ContractAddress = address
	WatchContractList[address] = contract
}

// 删除关注合约
func RemoveContract(address string) {
	address = strings.ToLower(address)
	delete(WatchContractList, address)
}

// 更新关注合约
func UpdateContract(name string, decimal int, address string) {
	address = strings.ToLower(address)
	if WatchContractList[address] != nil {
		return
	}

	WatchContractList[address].Decimal = decimal
	WatchContractList[address].Name = name
}
