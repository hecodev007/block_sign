package status

//冷热钱包状态
type WalletType string

const (
	WalletType_Cold WalletType = "cold"
	WalletType_Hot  WalletType = "hot"
)
