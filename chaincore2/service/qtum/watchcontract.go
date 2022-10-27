package qtum

import (
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: 合约地址
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {
	InsertContract("QC", 8, "f2033ede578e17fa6231047265010445bca8cf1c")
	InsertContract("HPY", 8, "f2703e93f87b846a7aacec1247beaec1c583daa4")
}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
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
	delete(WatchContractList, address)
}
