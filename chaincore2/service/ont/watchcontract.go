package ont

import (
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: 合约地址
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {
	InsertContract("ONT", 1, "0100000000000000000000000000000000000000")
	InsertContract("ONG", 9, "0200000000000000000000000000000000000000")
	InsertContract("SIN", 6, "6df81bc4b30189b0987b54f1d02b62f732cfd8a1")
	InsertContract("WING", 9, "00c59fcd27a562d6397883eab1f2fff56e58ef80")
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
