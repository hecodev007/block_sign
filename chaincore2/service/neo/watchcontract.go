package neo

import (
	"github.com/group-coldwallet/chaincore2/models"
	"strings"
)

// 关注地址列表, key: 合约地址
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

const NeoAssert = "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b"

// 初始化关注合约
func InitContract() {
	InsertContract("NEO", 0, "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b")
	//InsertContract("GAS", 8, "0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7")
}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

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
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	delete(WatchContractList, address)
}
