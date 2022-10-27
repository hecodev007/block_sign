package transpush

import (
	"database/sql"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
)

// 同步金额模型
type Sync struct {
	// nothing
}

func (r *Sync) Run(reqexit <-chan bool) {
	log.Debug("Run Sync")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("Sync exit", s)
			run = false
			break

		default:
			dispossSync()
			time.Sleep(time.Second * 5)
		}
	}
	WaitGroupTransPush.Done()
}

func dispossSync() {
	dbgroup := dao.TransPushGetDBEnginGroup()
	//$coin_set = $this->cache->get_set('coin_set');
	list := make([]*entity.FcTxClearDetail, 0)
	err := dbgroup.SQL("select id,mch_id,tx_id,tx_n,coin_type,dir,addr,amount,addr_type,from_tx_id from fc_tx_clear_detail where mch_id > 0 and addr_type > 0 and amount > 0 and is_over = 0").Limit(1000).Find(&list)
	if err != nil {
		log.Error(err)
		return
	}

	if len(list) == 0 {
		return
	}

	for _, vv := range list {
		fee := decimal.NewFromInt(0)
		txclear := &entity.FcTxClear{}
		isfind, err := dao.TransPushGet(txclear, "select tx_fee from fc_tx_clear where tx_id = ? and coin_type = ?", vv.TxId, vv.CoinType)
		if err != nil {
			log.Error(err)
			break
		}

		if isfind {
			fee, _ = decimal.NewFromString(txclear.TxFee)
		}

		session := dbgroup.NewSession()
		if session == nil {
			return
		}
		defer session.Close()
		session.Begin()

		amount := decimal.NewFromInt(0)
		if vv.Dir == 2 {
			tmpamount, _ := decimal.NewFromString(vv.Amount)
			if in_array(txclear.Coin, []interface{}{"eth", "algo", "pcx", "nas", "etc", "zvc", "atom", "seek", "mdu", "stg", "cocos", "kava", "luna","lunc", "klay", "gxc", "cds", "ong"}) {
				tmpamount = tmpamount.Add(fee)
				amount = tmpamount.Mul(decimal.NewFromInt(-1))
			} else {
				amount = tmpamount.Mul(decimal.NewFromInt(-1))
			}
		} else {
			amount, _ = decimal.NewFromString(vv.Amount)
		}

		addressAmount := &entity.FcAddressAmount{}
		isfind, err = session.SQL("select id, amount from fc_address_amount where coin_type = ? and address = ?", vv.CoinType, vv.Addr).Get(addressAmount)
		if err != nil {
			log.Error(err)
			session.Rollback()
			return
		}

		var execError error
		var res sql.Result
		if isfind {
			updateAmount, _ := decimal.NewFromString(addressAmount.Amount)
			updateAmount = updateAmount.Add(amount)
			res, execError = session.Exec("update fc_address_amount set amount = ? where coin_type = ? and address = ?", updateAmount.String(), vv.CoinType, vv.Addr)
			log.Debug(res)
		} else {
			coin_id := 0
			if global.CoinDecimal[vv.CoinType] != nil {
				coin_id = global.CoinDecimal[vv.CoinType].Id
			}
			res, execError = session.Exec("insert into fc_address_amount(coin_id, type, coin_type, address, amount, app_id) values(?, ?, ?, ?, ?, ?)", coin_id, vv.AddrType, vv.CoinType, vv.Addr, amount.String(), vv.MchId)
			log.Debug(res)
		}
		log.Debug(session.LastSQL())

		if vv.Dir == 2 && vv.FromTxId != "" {
			if strings.ToLower(vv.CoinType) == "btm" {
				transPush := &entity.FcTransPush{}
				isfind, err = session.SQL("select id from fc_trans_push where vout_id = ? and is_in = 1", vv.FromTxId).Get(transPush)
				if err != nil {
					log.Error(err)
					session.Rollback()
					return
				}
				if isfind {
					session.Exec("update fc_trans_push set is_spent = 1 where id = ?", transPush.Id)
					log.Debug("更新花费成功", transPush.Id)
				}
			} else {
				transPush := &entity.FcTransPush{}
				isfind, err = session.SQL("select id from fc_trans_push where transaction_id = ? and is_in = 1 and trx_n = ?", vv.FromTxId, vv.TxN).Get(transPush)
				if err != nil {
					log.Error(err)
					session.Rollback()
					return
				}
				if isfind {
					session.Exec("update fc_trans_push set is_spent = 1 where id = ?", transPush.Id)
					log.Debug("更新花费成功", transPush.Id)
				}
			}
		}

		if vv.Dir == 2 {
			addressAmount2 := &entity.FcAddressAmount{}
			isfind2, err := session.SQL("select id, pending_amount from fc_address_amount where coin_type = ? and address = ?", vv.CoinType, vv.Addr).Get(addressAmount2)
			if err != nil {
				log.Error(err)
				session.Rollback()
				return
			}
			if isfind2 {
				pending_amount, _ := decimal.NewFromString(addressAmount2.PendingAmount)
				pending_amount = pending_amount.Add(amount)
				if pending_amount.Cmp(decimal.NewFromInt(0)) > 0 && in_array(strings.ToLower(vv.CoinType), []interface{}{"btm", "stx", "omni-entc", "ckb"}) {
					res, execError = session.Exec("update fc_address_amount set pending_amount = ? where coin_type = ? and address = ?", pending_amount.String(), vv.CoinType, vv.Addr)
					log.Debug(res)
				}
			}
		}
		ref, err := session.Exec("update fc_tx_clear_detail set is_over = 1 where id = ?", vv.Id)

		refAffect, _ := ref.RowsAffected()
		resAffect, _ := res.RowsAffected()
		if err == nil && execError == nil && refAffect > 0 && resAffect > 0 {
			session.Commit()

			log.Debug("ok_", vv.Id)
			if vv.Dir == 2 {
				txclear := &entity.FcTxClear{}
				isfind3, err := dao.TransPushGet(txclear, "select id, coin, coin_type from fc_tx_clear where tx_id = ?", vv.TxId)
				if err != nil {
					log.Error(err)
					return
				}
				if isfind3 {
					if (in_array(strings.ToLower(txclear.Coin), []interface{}{"eth", "nas", "etc"}) && strings.ToLower(txclear.Coin) != strings.ToLower(txclear.CoinType)) || in_array(strings.ToLower(txclear.Coin), []interface{}{"bnb"}) {
						//Db::name("address_amount")->where(['coin_type'=>$tx['coin'], 'address'=>$vv['addr']])->setInc('amount', bcmul(-1, $fee));
						tmpfee := fee.Mul(decimal.NewFromInt(-1))
						dao.TransPushUpdate("update fc_address_amount set amount = amount + ? where coin_type = ? and address = ?", tmpfee.String(), txclear.Coin, vv.Addr)
					}
				}
			}
		} else {
			session.Rollback()
			log.Debug("error__", vv.Id)
		}
	}
}
