package vo

type Mtrbalance struct {
	CoinName     string `json:"coinName"`
	Decimal      int    `json:"decimal"`
	BalanceFloat string `json:"balanceFloat"`
}
