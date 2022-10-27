package base

import (
	"errors"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/shopspring/decimal"
)

const (
	BtcCoinName   = "btc" //币种名字
	BtcDecimalBit = 8     //币种精度
	//minFee     = 0.00001 //最小手续费
	//maxFee     = 0.0005  //最大手续费
	//utxoNumber = 15      //允许15个utxo
	//minAmount  = 0.0001  //最小金额过滤（只有该地址金额大于等于0.00001才会调用归集）
	//maxAmount  = 0.001   //最大金额过滤（只有该地址金额小于0.01才会调用归集）
)

type CoinConf struct {
	Env           string          `toml:"env"`             //开发环境
	MinFee        decimal.Decimal `toml:"min_fee"`         //最小手续费
	MaxFee        decimal.Decimal `toml:"max_fee"`         //最大手续费
	MinUtxoAmount decimal.Decimal `toml:"min_utxo_amount"` //utxo限制使用最小金额
	MaxUtxoAmount decimal.Decimal `toml:"max_utxo_amount"` // utxo限制使用最大金额
	MinAmount     decimal.Decimal `toml:"min_amount"`      //最小金额过滤（只有该地址金额大于等于0.00001才会调用归集）
	MaxAmount     decimal.Decimal `toml:"max_amount"`      //最大金额过滤（只有该地址金额小于0.01才会调用归集）
	UtxoNumber    uint            `toml:"utxo_number"`     //允许utxo数量
	AddrNumber    uint            `toml:"addr_number"`     //允许addr数量
	Second        uint            `toml:"second"`          //定时任务秒数
	HasCold       bool            `toml:"has_cold"`        //是否包含冷地址
}

func (c *CoinConf) New() {
	c = &CoinConf{
		Env:        "debug",
		MinFee:     decimal.NewFromFloat(0.00001), //最小手续费
		MaxFee:     decimal.NewFromFloat(0.0005),  //最大手续费
		MinAmount:  decimal.NewFromFloat(0.0001),  //最小金额过滤（只有该地址金额大于等于0.00001才会调用归集）
		MaxAmount:  decimal.NewFromFloat(0.01),    //最大金额过滤（只有该地址金额小于0.01才会调用归集）
		UtxoNumber: 15,                            //允许15个utxo
		Second:     120,
	}
}

func (c *CoinConf) Check() error {

	log.Infof("%+v", c)
	if c.MinFee.GreaterThan(decimal.NewFromFloat(0.001)) {
		return errors.New("最小手续费设置不允许大于0.001")
	}
	if c.MaxFee.GreaterThan(decimal.NewFromFloat(0.01)) {
		return errors.New("最大手续费设置不允许大于0.01")
	}
	if c.MinAmount.GreaterThan(decimal.NewFromFloat(10)) {
		return errors.New("归集条件，最小金额过滤设置不允许大于10")
	}
	if c.MaxAmount.GreaterThan(decimal.NewFromFloat(1000)) {
		return errors.New("归集条件，最大金额过滤设置不允许大于1000")
	}
	if c.UtxoNumber > 50 {
		return errors.New("utxo数量限制，不允许设置大于50")
	}
	if c.Second < 10 {
		return errors.New("秒数限制，不允许设置小于10")
	}
	return nil

}
