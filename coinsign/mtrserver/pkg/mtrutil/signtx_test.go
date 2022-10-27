package mtrutil

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/zwjlink/meterutil/tx"
	"math/big"
	"testing"
)

func TestMtrTpl_SignTxTpl(t *testing.T) {
	mtrTpl := &MtrTpl{
		ChainTag:     82,
		Expiration:   18,
		Gas:          21000,
		GasPriceCoef: 0,
		BlockRef:     715983,
		Outs: []ToTpl{
			{
				//ToAddress: "0x79c77f43ff0b291c1ae5d5e2aa1143949e4366fb",
				ToAddress: "0x932554c470604217fede2b56adbdaa3d89a24bf4",
				//ToAddress: "0x800803a96bdFD167B42A2Ef27F87e3b7996e019D",
				//0x26c0bd140e480902a0edbd2ac55e61bdb6106c7b9f6622e04ee834b636853493
				ToAmountInt: big.NewInt(decimal.NewFromFloat(0.02).Shift(18).IntPart()),
				Token:       tx.MeterToken,
				//Token: tx.MeterGovToken,
			},
		},
		//0x6A1ff04Be818B2372E2854Ad70D4D50274F208E9
		//0x70d74ffd764e11390d8ee31feff9aa4fef4bc1d44733396959d78739e0c34f71
		FromAddr:    "0x2e8e94e5b4ca3001f380fc0cd02fc803b24308d7",
		FromPrivKey: "0x23288e2b4ea5b279c564ce9ad7a1aa57e02064aa9eb3350abe4289852559fab2",

		//0xbc505934fe03AD4A141dB74855a81C33e70Cf99f
		//0x5ed9a20de12173b2d2295c70700c62ea99c370165e9f89728ca13dcabb50bc02
		//FeePayerPrivKey: "0x0129e0218b2246ea6103c73a7680234afa41afac734631780e00868cc267b6ef",
	}
	raw, err := mtrTpl.SignTxTpl()
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("raw tx :%s", raw)

}

func TestDecode(t *testing.T) {

	in, _ := hexutil.DecodeBig("0x43cbe88b679ee680000")
	in2, _ := hexutil.DecodeBig("0x43cbe88b679ee680000")
	t.Log(in)
	t.Log(in2)

	mtrBalanceFloat := decimal.NewFromBigInt(in, -18)
	mtrBalanceFloat = mtrBalanceFloat.Add(decimal.NewFromFloat(0.1))
	t.Log(mtrBalanceFloat)
}
