package global

import (
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
)

var TransferModel map[string]model.TransferModel

func InitTransferModel() {
	//币种交易模型（非币种结构模型，比如BTM需要走账户模型交易流程）
	tfModel := make(map[string]model.TransferModel, 0)
	for _, v := range conf.Cfg.TransferModel.UtxoModel {
		tfModel[v] = model.TransferModelUtxo
	}
	for _, v := range conf.Cfg.TransferModel.AccountModel {
		tfModel[v] = model.TransferModelAccount
	}
	TransferModel = tfModel
	log.Info("加载币种交易模型")
}
