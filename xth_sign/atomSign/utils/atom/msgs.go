package atom

//
//import (
//	sdk "github.com/cosmos/cosmos-sdk/types"
//)
//
//const RouterKey = "bank"
//type MsgSends []sdk.Msg
//
//type MsgSend struct {
//	FromAddress sdk.AccAddress `json:"from_address" yaml:"from_address"`
//	ToAddress   sdk.AccAddress `json:"to_address" yaml:"to_address"`
//	Amount      sdk.Coins      `json:"amount" yaml:"amount"`
//}
//
//// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
//func NewMsgSend(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) MsgSend {
//	return MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
//}
//
//// Route Implements Msg.
//func (msg MsgSend) Route() string { return RouterKey }
//
//// Type Implements Msg.
//func (msg MsgSend) Type() string { return "send" }
//
//// ValidateBasic Implements Msg.
//func (msg MsgSend) ValidateBasic() sdk.Error {
//	if msg.FromAddress.Empty() {
//		return sdk.ErrInvalidAddress("missing sender address")
//	}
//	if msg.ToAddress.Empty() {
//		return sdk.ErrInvalidAddress("missing recipient address")
//	}
//	if !msg.Amount.IsValid() {
//		return sdk.ErrInvalidCoins("send amount is invalid: " + msg.Amount.String())
//	}
//	if !msg.Amount.IsAllPositive() {
//		return sdk.ErrInsufficientCoins("send amount must be positive")
//	}
//	return nil
//}
//
//// GetSignBytes Implements Msg.
//func (msg MsgSend) GetSignBytes() []byte {
//	return sdk.MustSortJSON(atomCdc.MustMarshalJSON(msg))
//}
//
//// GetSigners Implements Msg.
//func (msg MsgSend) GetSigners() []sdk.AccAddress {
//	return []sdk.AccAddress{msg.FromAddress}
//}
