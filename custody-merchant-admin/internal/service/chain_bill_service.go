package service

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/domain/dto"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/model/chainBill"
	"custody-merchant-admin/module/dict"
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"strconv"
	"time"
)

func CreateChainBill(bill *chainBill.Entity) error {
	err := bill.CreateChainBill()
	if err != nil {
		return err
	}
	return nil
}

func UpdateChainBillByMap(id int64, mp map[string]interface{}) error {
	cbDao := chainBill.NewEntity()
	err := cbDao.UpdatesChainBill(id, mp)
	if err != nil {
		return err
	}
	return nil
}

func FindChainBillService(info *domain.ChainBillSelect) ([]domain.ChainBillInfo, int64, error) {
	list, total, err := deals.FindChainBillList(info)
	if err != nil {
		return list, total, err
	}
	return list, total, nil
}

func FindChainBillExport(info *domain.ChainBillSelect) (bytes.Buffer, error) {
	info.Offset = 0
	info.Limit = 99999
	bill, _, err := deals.FindChainBillList(info)
	if err != nil {
		return bytes.Buffer{}, err
	}
	xFile := xlsx.NewFile()
	sheet, err := xFile.AddSheet("Sheet1")
	if err != nil {
		return bytes.Buffer{}, err
	}
	title := []string{"序号", "商户ID", "账户状态", "手机号", "业务线ID", "业务线名称", "主链币", "代币名", "订单ID", "订单类型",
		"订单状态", "发送地址", "接收地址", "MEMO", "交易数量", "手续费", "矿工费", "销毁数量", "区块确认数", "TXID", "区块高度",
		"钱包处理", "是否冷钱包", "冷钱包处理状态", "冷钱包处理结果", "是否回滚", "TXID时间", "创建时间", "确认时间"}
	r := sheet.AddRow()
	var ce *xlsx.Cell
	for _, v := range title {
		ce = r.AddCell()
		ce.Value = v
	}
	for i := 0; i < len(bill); i++ {
		r = sheet.AddRow()
		// 序号
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(bill[i].Id, 10)
		// 商户ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].MerchantId)
		// 账户状态
		ce = r.AddCell()
		ce.Value = dict.IsTestText[bill[i].IsTest]
		// 手机号
		ce = r.AddCell()
		ce.Value = bill[i].Phone
		// 业务线ID
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].ServiceId)
		// 业务线名
		ce = r.AddCell()
		ce.Value = bill[i].ServiceName
		// 主链币
		ce = r.AddCell()
		ce.Value = bill[i].ChainName
		// 代币
		ce = r.AddCell()
		ce.Value = bill[i].CoinName
		// 订单ID
		ce = r.AddCell()
		ce.Value = bill[i].SerialNo
		// 订单类型
		ce = r.AddCell()
		ce.Value = dict.TxTypeNameList[bill[i].TxType]
		// 订单状态
		ce = r.AddCell()
		ce.Value = dict.BillStateList[bill[i].BillStatus]
		// 发送地址
		ce = r.AddCell()
		ce.Value = bill[i].TxFromAddr
		// 接收地址
		ce = r.AddCell()
		ce.Value = bill[i].TxToAddr
		// Memo
		ce = r.AddCell()
		ce.Value = bill[i].Memo
		// 数量
		ce = r.AddCell()
		ce.Value = bill[i].Nums.String()
		// 手续费
		ce = r.AddCell()
		ce.Value = bill[i].Fee.String()
		// 矿工费
		ce = r.AddCell()
		ce.Value = bill[i].BurnFee.String()
		// 销毁费
		ce = r.AddCell()
		ce.Value = bill[i].DestroyFee.String()
		// 区块确认数
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].ConfirmNums)
		// TXID
		ce = r.AddCell()
		ce.Value = bill[i].TxId
		// 区块高度
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", bill[i].Height)
		// 钱包处理
		ce = r.AddCell()
		ce.Value = bill[i].IsWalletDealName
		// 是否冷钱包
		ce = r.AddCell()
		ce.Value = bill[i].IsColdWalletName
		// 冷钱包处理状态
		ce = r.AddCell()
		ce.Value = bill[i].ColdWalletStateName
		// 冷钱包处理结果
		ce = r.AddCell()
		ce.Value = bill[i].ColdWalletResultName
		// 是否回滚
		ce = r.AddCell()
		ce.Value = bill[i].IsRebackName
		// 交易时间
		ce = r.AddCell()
		ce.Value = bill[i].TxTime
		// 创建时间
		ce = r.AddCell()
		ce.Value = bill[i].CreateTime
		// 确认时间
		ce = r.AddCell()
		ce.Value = bill[i].ConfirmTime
	}
	//将数据存入buff中
	var buff bytes.Buffer
	if err = xFile.Write(&buff); err != nil {
		return bytes.Buffer{}, err
	}
	return buff, nil
}

// RollBackChainBill
// 回滚订单
func RollBackChainBill(id int64) error {
	// 先获取链上订单
	dao := chainBill.NewEntity()
	err := dao.FindChainBillById(id)
	if err != nil {
		return err
	}
	// TODO 调用钱包回滚接口，回滚数据
	back, err := OrderRollBack(&domain.OrderOperateReq{
		OutOrderId: dao.SerialNo,
		BusinessId: int64(dao.ServiceId),
		AccountId:  int(dao.MerchantId),
	})
	if err != nil {
		return err
	}
	if back == 0 {
		return errors.New("无法回滚链上账单")
	}
	// 资产回滚
	// 修改账单状态
	err = RollbackBillAssets(&domain.BillInfo{
		SerialNo: dao.SerialNo,
	})
	if err != nil {
		return err
	}
	// 修改链上订单状态
	mp := map[string]interface{}{
		"bill_status":        5,
		"tx_type":            5,
		"is_reback":          1,
		"cold_wallet_result": 1,
		"updated_at":         time.Now().Local(),
	}
	err = dao.UpdatesChainBill(dao.Id, mp)
	if err != nil {
		return err
	}
	return nil
}

// RePushChainBill
// 重推链上订单
func RePushChainBill(id int64) error {
	// 先获取链上订单
	dao := chainBill.NewEntity()
	err := dao.FindChainBillById(id)
	if err != nil {
		return err
	}
	// 调用钱包重推接口
	return nil
}

func FindChainBillNoUpChain(bill domain.BillInfo) ([]chainBill.Entity, error) {
	cbDao := dto.GetChainBillInfo(bill)
	return cbDao.FindChainBillNoUpChain()
}

func FindChainBillNoUpChainTxid(txid string) (chainBill.Entity, error) {
	cbDao := chainBill.NewEntity()
	return cbDao.FindChainBillNoUpChainTxid(txid)
}

func FindChainBillChainTxid(txid string) (chainBill.Entity, error) {
	cbDao := chainBill.NewEntity()
	return cbDao.FindChainBillChainTxid(txid)
}

func FindChainBillChainSerialNo(serialNo string) (chainBill.Entity, error) {
	cbDao := chainBill.NewEntity()
	return cbDao.FindChainBillChainSerialNo(serialNo)
}

func FindChainBillNoUpChainSerialNo(serialNo string) (chainBill.Entity, error) {
	cbDao := chainBill.NewEntity()
	return cbDao.FindChainBillNoUpChainSerialNo(serialNo)
}
