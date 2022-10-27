package domain

import "github.com/shopspring/decimal"

type RingObj struct {
	RingInfo  []RingInfo  `json:"ring_info"`
	RingItems []RingItems `json:"ring_items"`
}

type RingInfo struct {
	Value     decimal.Decimal `json:"value"`
	Name      string          `json:"name"`
	ItemStyle ItemStyle       `json:"itemStyle"`
}

type RingItems struct {
	Value decimal.Decimal `json:"value"`
	Name  string          `json:"name"`
	Color string          `json:"color"`
}

type ItemStyle struct {
	Color string `json:"color"`
}
