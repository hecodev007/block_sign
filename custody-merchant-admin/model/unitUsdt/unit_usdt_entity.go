package unitUsdt

import "github.com/shopspring/decimal"

type UnitUsdt struct {
	Id    int             `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Name  string          `json:"name" gorm:"column:name"`
	Ratio decimal.Decimal `json:"ratio" gorm:"column:ratio"`
}

func (unit *UnitUsdt) TableName() string {
	return "unit_usdt"
}
