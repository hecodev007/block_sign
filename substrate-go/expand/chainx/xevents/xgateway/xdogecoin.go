package xgateway

/// XGatewayDogecoin Type
type XGatewayDogecoin struct {
	XGatewayDogecoin_HeaderInserted              []EventXGatewayBitcoinHeaderInserted
	XGatewayDogecoin_TxProcessed                 []EventXGatewayBitcoinTxProcessed
	XGatewayDogecoin_Deposited                   []EventXGatewayBitcoinDeposited
	XGatewayDogecoin_Withdrawn                   []EventXGatewayBitcoinWithdrawn
	XGatewayDogecoin_UnclaimedDeposit            []EventXGatewayBitcoinUnclaimedDeposit
	XGatewayDogecoin_PendingDepositRemoved       []EventXGatewayBitcoinPendingDepositRemoved
	XGatewayDogecoin_WithdrawalProposalCreated   []EventXGatewayBitcoinWithdrawalProposalCreated
	XGatewayDogecoin_WithdrawalProposalVoted     []EventXGatewayBitcoinWithdrawalProposalVoted
	XGatewayDogecoin_WithdrawalProposalDropped   []EventXGatewayBitcoinWithdrawalProposalDropped
	XGatewayDogecoin_WithdrawalProposalCompleted []EventXGatewayBitcoinWithdrawalProposalCompleted
	XGatewayDogecoin_WithdrawalFatalErr          []EventXGatewayBitcoinWithdrawalFatalErr
}

