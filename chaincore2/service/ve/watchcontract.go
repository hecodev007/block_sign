package ve

import (
	"github.com/group-coldwallet/chaincore2/models"
)

// 关注地址列表, key: 合约名称
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {
	InsertContract("vtho", 18, "0x0000000000000000000000000000456e65726779")
}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	if WatchContractList[name] != nil {
		return
	}

	contract := new(models.ContractInfo)
	contract.Name = name
	contract.Decimal = decimal
	contract.ContractAddress = address
	WatchContractList[name] = contract
}

// 删除关注合约
func RemoveContract(name string) {
	delete(WatchContractList, name)
}

func HasContact(addr string) (bool,*models.ContractInfo) {
	for _,v := range WatchContractList {
		if v.ContractAddress == addr {
			return true,v
		}
	}
	return false,nil
}