package dogeutil

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestSignTxTpl(t *testing.T) {
	tpl := &DogeTxTpl{
		TxIns: []DogeTxInTpl{
			DogeTxInTpl{
				FromAddr:        "DKLzbyszeXtRifSe4LZEJKonFwUgFUo3hM",
				FromPrivkey:     "QU7mLQc6BX1L1eqCFt4sMPoJ8Cgh1AMmxWZo5k7piKEdNxhnYKhs",
				FromTxid:        "082665e991487c770a370c893efbbd6c867a65e322c6df38697092261d6d047a",
				FromIndex:       uint32(0),
				FromAmountInt64: decimal.NewFromFloat(200.00000000).Shift(dogeDecimal).IntPart(),
			},
		},
		TxOuts: []DogeTxOutTpl{
			DogeTxOutTpl{
				ToAddr:        "D8AD6b5poKDqPHYHn2t7bpArUkNYhEzPhK",
				ToAmountInt64: decimal.NewFromFloat(20).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "D8AD6b5poKDqPHYHn2t7bpArUkNYhEzPhK",
				ToAmountInt64: decimal.NewFromFloat(30).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "D8AD6b5poKDqPHYHn2t7bpArUkNYhEzPhK",
				ToAmountInt64: decimal.NewFromFloat(40).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "DGEFyZdVb5gxmd9csdZ66xSzzEdKjWT8C3",
				ToAmountInt64: decimal.NewFromFloat(10).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "DGEFyZdVb5gxmd9csdZ66xSzzEdKjWT8C3",
				ToAmountInt64: decimal.NewFromFloat(20).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "DGEFyZdVb5gxmd9csdZ66xSzzEdKjWT8C3",
				ToAmountInt64: decimal.NewFromFloat(30).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "D7o1yTnXxhbyS9Kq65f1JGNdCdg8CRjaji",
				ToAmountInt64: decimal.NewFromFloat(15).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "D7o1yTnXxhbyS9Kq65f1JGNdCdg8CRjaji",
				ToAmountInt64: decimal.NewFromFloat(25).Shift(dogeDecimal).IntPart(),
			},
			DogeTxOutTpl{
				ToAddr:        "D7o1yTnXxhbyS9Kq65f1JGNdCdg8CRjaji",
				ToAmountInt64: decimal.NewFromFloat(7).Shift(dogeDecimal).IntPart(),
			},
		},
	}
	t.Log(SignTxTpl(tpl))
}
