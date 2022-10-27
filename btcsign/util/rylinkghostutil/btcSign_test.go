package rylinkghostutil

import (
	"github.com/group-coldwallet/btcsign/model/bo"
	"testing"
)

func TestSignTxTpl(t *testing.T) {
	initMain()
	tpl := &bo.BtcTxTpl{
		TxIns: []bo.BtcTxInTpl{
			bo.BtcTxInTpl{
				FromAddr:    "GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG",
				FromPrivkey: "RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx",
				FromTxid:    "9d69514242fd3a53aa2f55125c70aa7fe3e60462295476d0c2704d0293f90ec8",
				FromIndex:   uint32(0),
				FromAmount:  int64(100000000),
			},
			//bo.BtcTxInTpl{
			//	FromAddr:    "339vF2xsau46qVty8stqmhgT5jr9aiu92j",
			//	FromPrivkey: "KxBie4tbASZ4LiFb37jLDRf6gmgEuQ71WFqXR4PqXuszCoo6yigN",
			//	FromTxid:    "7aa6459f36c8358fdc1a6fcb3eec3d8f1112a84699f82ec0c0699e0cf7d3882b",
			//	FromIndex:   uint32(0),
			//	FromAmount:  int64(12220),
			//},
			//bo.BtcTxInTpl{
			//	FromAddr:    "339vF2xsau46qVty8stqmhgT5jr9aiu92j",
			//	FromPrivkey: "KxBie4tbASZ4LiFb37jLDRf6gmgEuQ71WFqXR4PqXuszCoo6yigN",
			//	FromTxid:    "34331626dfe11dbf803d181e6339d0cd0a8a27fe4ff9dedf66ef20d68fb255ab",
			//	FromIndex:   uint32(0),
			//	FromAmount:  int64(23500),
			//},
		},
		TxOuts: []bo.BtcTxOutTpl{
			bo.BtcTxOutTpl{
				ToAddr:   "GeazzFGwzRcLminwTrWxfXxLc7Ga8TMsUG",
				ToAmount: int64(90000000),
			},
			bo.BtcTxOutTpl{
				ToAddr:   "GdVufw2QtMTd2fddoFMMKQq4HyydARsGNX",
				ToAmount: int64(9500000),
			},
		},
	}

	//10个输出
	//for i := 0; i < 3; i++ {
	//	tpl.TxOuts = append(tpl.TxOuts, bo.BtcTxOutTpl{
	//		ToAddr:   "32Yas4gZL9kettai9PiMg1HzRtpf44YmTH",
	//		ToAmount: 15326,
	//	})
	//}

	//10个输出
	//for i := 0; i < 10; i++ {
	//	tpl.TxOuts = append(tpl.TxOuts, bo.BtcTxOutTpl{
	//		ToAddr:   "38Uw23x5G5pgmQL9pJDUfacJrwzpXKqr7S",
	//		ToAmount: 12000,
	//	})
	//}
	t.Log(SignTxTpl(tpl))
}
