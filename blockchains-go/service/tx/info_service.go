package tx

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"strings"
)

type TxInfoService struct {
}

func (t *TxInfoService) FindTxList(params *model.SearchTx) (results []*model.TxInfo, count int64, err error) {
	var (
		mchInfo *entity.FcMch
		datas   []*entity.FcTxTransaction
	)

	account := []string{"eth", "eos", "algo", "hx", "usdt", "etc", "xrp", "nas", "bnb", "wbc", "qc", "zvc", "seek", "mdu", "stg"}
	accountMap := make(map[string]struct{})
	for _, v := range account {
		accountMap[v] = struct{}{}
	}
	err = params.Check()
	if err != nil {
		return nil, 0, err
	}
	mchInfo, err = dao.FcMchFindByPlatform(params.Sfrom)
	if err != nil {
		return nil, 0, err
	}
	datas, count, err = dao.FcTxTransactionFindList(mchInfo.Id, params.Type, strings.ToLower(params.CoinName), params.DateStart, params.DateEnd)
	if err != nil {
		return nil, 0, err
	}
	if count != 0 {
		for _, v := range datas {
			amount := decimal.Zero
			if v.TxType == 2 || v.TxType == 11 {
				if _, ok := accountMap[strings.ToLower(v.CoinType)]; ok {
					//非账户模型需要扣除手续费
					am, _ := decimal.NewFromString(v.Amount)
					fee, _ := decimal.NewFromString(v.TxFee)
					amount = am.Sub(fee)
				}
			}

			results = append(results, &model.TxInfo{
				CoinType:     v.CoinType,
				OuterOrderNo: v.OuterOrderNo,
				OrderNo:      v.OrderNo,
				BlockHeight:  v.BlockHeight,
				Timestamp:    v.Timestamp,
				TxId:         v.TxId,
				TxType:       v.TxType,
				FromAddress:  v.FromAddress,
				ToAddress:    v.ToAddress,
				Memo:         v.Memo,
				Amount:       amount.String(),
				TxFee:        v.TxFee,
				TxFeeCoin:    v.TxFeeCoin,
				ContrastTime: v.ContrastTime,
			})
		}
	}
	return results, count, nil

}

func NewTxInfoService() service.TransactionInfoService {
	return &TxInfoService{}
}
