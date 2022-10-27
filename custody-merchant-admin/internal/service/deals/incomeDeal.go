package deals

import (
	"bytes"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model/incomeAccount"
	"fmt"
	"github.com/tealeg/xlsx"
	"strconv"
)

func FindIncomePage(search *domain.SearchIncome) (domain.IncomeList, int64, error) {
	dao := incomeAccount.NewEntity()
	res := domain.IncomeList{}
	page, count, err := dao.FindPage(search)
	if err != nil {
		return domain.IncomeList{}, count, err
	}
	for i, _ := range page {
		p := page[i]
		p.Serial = int64(i + 1)
		res.List = append(res.List, p)
	}
	return res, count, err
}
func FindIncomeChart(search *domain.SearchIncome) (domain.IncomeList, error) {
	dao := incomeAccount.NewEntity()
	res := domain.IncomeList{}
	charts, err := dao.FindChart(search)
	if err != nil {
		return res, err
	}
	for i := 0; i < len(charts); i++ {
		res.TopUp = res.TopUp.Add(charts[i].TopUpIncome)
		res.Withdraw = res.Withdraw.Add(charts[i].WithdrawIncome)
		res.Combo = res.Combo.Add(charts[i].ComboIncome)
		// 总收益 = 总收益+套餐收益
		res.Totals = res.Totals.Add(charts[i].TotalIncome)
		if i < 5 {
			res.TopList = append(res.TopList, domain.TopInfo{
				CoinName: charts[i].CoinName,
				CoinId:   charts[i].CoinId,
				Price:    charts[i].TotalIncome,
			})
		}
	}
	return res, err
}

func FindIncomeExcelExport(search *domain.SearchIncome) (bytes.Buffer, error) {
	dao := incomeAccount.NewEntity()
	search.Offset = 0
	search.Limit = 10000
	page, _, err := dao.FindPage(search)
	xFile := xlsx.NewFile()
	sheet, err := xFile.AddSheet("Sheet1")
	if err != nil {
		return bytes.Buffer{}, err
	}
	title := []string{"序号", "姓名", "商户ID", "手机号", "邮箱", "套餐收费类型", "套餐收费模式", "业务线ID", "业务线名称", "主链币",
		"代币", "充值笔数", "充值金额", "充值手续费", "销毁数量", "充值收益", "提现笔数", "提现金额", "提现手续费", "矿工费", "销毁数量",
		"提现收益", "套餐收益", "总收益"}
	r := sheet.AddRow()
	var ce *xlsx.Cell
	for _, v := range title {
		ce = r.AddCell()
		ce.Value = v
	}
	for i := 0; i < len(page); i++ {
		r = sheet.AddRow()
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(int64(i+1), 10) // 序号
		ce = r.AddCell()
		ce.Value = page[i].UserName // 姓名
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(page[i].MerchantId, 10) // 商户ID
		ce = r.AddCell()
		ce.Value = page[i].UserPhone // 手机号
		ce = r.AddCell()
		ce.Value = page[i].UserEmail // 邮箱
		ce = r.AddCell()
		ce.Value = page[i].ComboTypeName // 套餐收费类型
		ce = r.AddCell()
		ce.Value = page[i].ComboModelName // 套餐收费模式
		ce = r.AddCell()
		ce.Value = strconv.FormatInt(int64(page[i].ServiceId), 10) // 业务线ID
		ce = r.AddCell()
		ce.Value = page[i].ServiceName // 业务线名称
		ce = r.AddCell()
		ce.Value = page[i].ChainName // 主链币
		ce = r.AddCell()
		ce.Value = page[i].CoinName // 代币
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", page[i].TopUpNums) // 充值笔数
		ce = r.AddCell()
		ce.Value = page[i].TopUpPrice.String() // 充值金额
		ce = r.AddCell()
		ce.Value = page[i].ToUpFee.String() // 充值手续费
		ce = r.AddCell()
		ce.Value = page[i].ToUpDestroy.String() // 销毁数量
		ce = r.AddCell()
		ce.Value = page[i].TopUpIncome.String() // 充值收益户
		ce = r.AddCell()
		ce.Value = fmt.Sprintf("%d", page[i].WithdrawNums) // 提现笔数
		ce = r.AddCell()
		ce.Value = page[i].WithdrawPrice.String() // 提现金额
		ce = r.AddCell()
		ce.Value = page[i].WithdrawFee.String() // 提现手续费
		ce = r.AddCell()
		ce.Value = page[i].MinerFee.String() // 矿工费
		ce = r.AddCell()
		ce.Value = page[i].WithdrawDestroy.String() // 销毁数量
		ce = r.AddCell()
		ce.Value = page[i].WithdrawIncome.String() // 提现收益户
		ce = r.AddCell()
		ce.Value = page[i].ComboIncome.String() // 套餐收益户
		ce = r.AddCell()
		ce.Value = page[i].TotalIncome.String() // 总收益户
	}
	//将数据存入buff中
	var buff bytes.Buffer
	if err = xFile.Write(&buff); err != nil {
		return bytes.Buffer{}, err
	}

	return buff, nil
}
