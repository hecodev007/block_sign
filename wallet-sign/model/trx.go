package model

type TrxTransferParams struct {
	OrderRequestHead
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	FeeLimit        int64  `json:"fee_limit"`
	ContractAddress string `json:"contract_address"` //用于trc20转账
	AssetId         string `json:"asset_id"`         //用于trc10转账
}

type OrderRequestHead struct {
	OuterOrderNo string `json:"outer_order_no,omitempty"`
}
