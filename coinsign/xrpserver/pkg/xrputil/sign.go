package xrputil

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-chain/go-xrp/crypto"
	"github.com/go-chain/go-xrp/data"
	"github.com/shopspring/decimal"
	"strings"
)

type XprSignTpl struct {
	From         string //from地址
	FromPrivate  string //from地址私钥
	To           string //to 地址
	AmountFloat  string //发送金额
	FeeFloat     string //手续费
	FromSequence uint32 //from地址序列 外部使用account_info方法获取
	LastSequence uint32 //过期序列 外部使用rpc ：server_state方法获取并且加100上去
	Tag          uint32 //tag标记，只能数字
	Currency     string //货币类型 默认 XRP    其他币种尚未测试
	CoinDecimal  int32  //币种精度
}

//只校验xrp币种，对于XRP分类帐中非XRP货币金额精度暂时不做处理
//暂时不支持x开头地址。尚未测试
func (tpl *XprSignTpl) CheckParams() error {
	if tpl.From == "" {
		return errors.New("miss from")
	}
	if tpl.To == "" {
		return errors.New("miss to")
	}
	if tpl.Currency != "XRP" {
		return errors.New("目前只支持XRP，其他币种尚未测试")
	}
	if tpl.CoinDecimal <= 0 {
		return errors.New("miss coinDecimal")
	}

	if tpl.FromPrivate == "" {
		return errors.New("miss private key")
	}
	_, err := hex.DecodeString(tpl.FromPrivate)
	if err != nil {
		return fmt.Errorf("privkey format error:%s", err.Error())
	}

	//校验地址合法性   或者 base58.CheckDecode()
	//注意解码之后对比原地址是否相等，NewAccountFromAddress强制使用了0序列进行转换
	_, err = data.NewAccountFromAddress(tpl.From)
	if err != nil {
		return fmt.Errorf("from format error:%s", err.Error())
	}
	actTo, err := data.NewAccountFromAddress(tpl.To)
	//强制r开头,暂时忽略x开头地址
	if err != nil || !strings.HasPrefix(tpl.To, "r") {
		return fmt.Errorf("to format error:%s", err.Error())
	}
	if actTo.String() != tpl.To {
		return fmt.Errorf("to format error:%s  format:%s", tpl.To, actTo.String())
	}
	_, err = hex.DecodeString(tpl.FromPrivate)
	if err != nil {
		return fmt.Errorf("to format error:%s", err.Error())
	}

	amount, err := decimal.NewFromString(tpl.AmountFloat)
	if err != nil {
		return fmt.Errorf("amount format error:%s", err.Error())
	}
	if amount.Exponent() < -tpl.CoinDecimal {
		return fmt.Errorf("toamount decimal error:%d", amount.Exponent())
	}
	//xpr发送金额不能小于0.01，否则会返回本地签名认证失败。其他币种位置，暂时按xrp处理
	if amount.LessThan(decimal.NewFromFloat(0.001)) {
		return fmt.Errorf("toamount min 0.001,now:%s", amount.String())
	}
	fee, err := decimal.NewFromString(tpl.FeeFloat)
	if err != nil {
		return fmt.Errorf("fee format error:%s", err.Error())
	}
	if fee.Exponent() < -tpl.CoinDecimal {
		return fmt.Errorf("fee decimal error:%d", fee.Exponent())
	}

	//限制一下最小手续费 预防出不去
	if fee.LessThan(decimal.NewFromFloat(0.00001)) {
		return fmt.Errorf("fee min 0.001,now:%s", amount.String())
	}
	//限制一下最大手续费,预防支付过多
	if fee.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("fee max 1,now:%s", fee.String())
	}
	return nil
}

//瑞博币的单位为：1000000
//发送金额不能小于0.01，否则会返回本地签名认证失败。
//新的瑞波币地址必须转入20个Xrp才可以激活账户。
//交易后地址余额必须大于20，否则交易无法被确认，会被节点直接打回
func (tpl *XprSignTpl) XrpSignTx() (rawtx string, err error) {
	var (
		fromAccount *data.Account //发送账号
		toAccount   *data.Account //接收账号
		amount      *data.Amount  //发送金额
		fee         *data.Value   //发送手续费
		rawtxByte   []byte        //广播内容字节码
	)

	//检查参数
	err = tpl.CheckParams()
	if err != nil {
		return "", err
	}

	//基本参数赋值
	fromAccount, _ = data.NewAccountFromAddress(tpl.From)
	toAccount, _ = data.NewAccountFromAddress(tpl.To)
	pri, _ := hex.DecodeString(tpl.FromPrivate)
	key := crypto.LoadECDSKey(pri)

	//金额转换
	value := tpl.AmountFloat
	if tpl.Currency != "" {
		value += "/" + tpl.Currency
	}
	amount, _ = data.NewAmount(value)

	//这里很奇怪不能使用native == false转换，会换异常，还没研究。
	//fee, _ = data.NewValue(tpl.FeeFloat, false)
	feeFloatD, _ := decimal.NewFromString(tpl.FeeFloat)
	//xpr手续费精度取整
	fee, _ = data.NewValue(feeFloatD.Shift(6).String(), true)
	//根据Ripple官方文档，Ripple交易中Flags字段不是「2147483648」 (十六进制：0x80000000) 值的交易可能存在安全风险。
	flags := data.TransactionFlag(2147483648)
	txnBase := data.TxBase{
		TransactionType:    data.PAYMENT,
		Account:            *fromAccount,
		Sequence:           tpl.FromSequence,
		Fee:                *fee,
		LastLedgerSequence: &tpl.LastSequence,
		Flags:              &flags,
		SourceTag:          &tpl.Tag,
	}
	payment := &data.Payment{
		TxBase:         txnBase,
		Destination:    *toAccount,
		Amount:         *amount,
		DestinationTag: &tpl.Tag,
	}
	err = data.Sign(payment, key, nil)
	if err != nil {
		return "", err
	}
	_, rawtxByte, err = data.Raw(data.Transaction(payment))
	if err != nil {
		return "", err
	}
	rawtx = fmt.Sprintf("%X", rawtxByte)
	return
}
