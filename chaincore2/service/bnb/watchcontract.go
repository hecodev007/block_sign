package bnb

import (
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
)

// 关注地址列表, key: 合约名称
var WatchContractList map[string]*models.ContractInfo = make(map[string]*models.ContractInfo)

// 初始化关注合约
func InitContract() {
	InsertContract("BNB", 8, "bnb1ultyhpw2p2ktvr68swz56570lgj2rdsadq3ym2")
	InsertContract("TROY-9B8", 8, "bnb1scrark2sv6fpngyqxrryw9hw7y05euwntz45ae")
	InsertContract("TWT-8C2", 8, "bnb1fkyxlq9kz5368ux29aeeztslclgf8e7tja345x")
	InsertContract("AWC-986", 8, "bnb1g5xj69c0s0x646hug7j3vr6eamlkf7jw3cr3yw")
	InsertContract("RUNE-B1A", 8, "bnb1e4q8whcufp6d72w8nwmpuhxd96r4n0fstegyuy")
	InsertContract("CAN-677", 8, "bnb16w59lfh4y2cqvu8f7yr000ll37ldh4w6hnz7l0")
	InsertContract("AVA-645", 8, "bnb1dm9c7gccgd07td5r69m50u8fg8danfgqvlhj6c")
	//2020年07月10日15:59:09
	InsertContract("BIDR-0E9", 8, "bnb16w59lfh4y2cqvu8f7yr000ll37ldh4w6hnz7l0")
	InsertContract("BKRW-AB7", 8, "bnb18kha55gvsxl7gkdh8y329hu3p6wndh6jkwqnxn")
	InsertContract("SWINGBY-888", 8, "bnb1thagrtfude74x2j2wuknhj2savucy2tx0k58y9")
	InsertContract("BUSD-BD1", 8, "bnb19v2ayq6k6e5x6ny3jdutdm6kpqn3n6mxheegvj")
}

// 插入关注合约
func InsertContract(name string, decimal int, address string) {
	if WatchContractList[name] != nil {
		return
	}
	log.Debugf("合约：%s", name)
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
