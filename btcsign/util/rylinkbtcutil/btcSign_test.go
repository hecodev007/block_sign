package rylinkbtcutil

import (
	"github.com/group-coldwallet/btcsign/model/bo"
	"testing"
)

//3NHoHpdCTcex1LUTEyXzo22HpkhS3fJhP6,L2vbwFHd4jgucVu62tmxgmBBZPeXoyW42qesZxE3d9j97mH6GYjj

func TestSignTxTpl(t *testing.T) {
	tpl := &bo.BtcTxTpl{
		TxIns: []bo.BtcTxInTpl{
			bo.BtcTxInTpl{
				FromAddr:    "3NHoHpdCTcex1LUTEyXzo22HpkhS3fJhP6",
				FromPrivkey: "L2vbwFHd4jgucVu62tmxgmBBZPeXoyW42qesZxE3d9j97mH6GYjj",
				FromTxid:    "91c768f02afa9ca6245401509cfa1a8ea4931c4ebd7b663768d74e368f8e7553",
				FromIndex:   uint32(0),
				FromAmount:  int64(100000),
				//0.00100000
			},
			bo.BtcTxInTpl{
				FromAddr:    "3NHoHpdCTcex1LUTEyXzo22HpkhS3fJhP6",
				FromPrivkey: "L2vbwFHd4jgucVu62tmxgmBBZPeXoyW42qesZxE3d9j97mH6GYjj",
				FromTxid:    "8c965308d2a247d25ea8c0a19ff6597ee1da25d3a012c069bec703ff4fc307a2",
				FromIndex:   uint32(2),
				FromAmount:  int64(10000),
				//0.00100000
			},
			bo.BtcTxInTpl{
				FromAddr:    "3NHoHpdCTcex1LUTEyXzo22HpkhS3fJhP6",
				FromPrivkey: "L2vbwFHd4jgucVu62tmxgmBBZPeXoyW42qesZxE3d9j97mH6GYjj",
				FromTxid:    "8c965308d2a247d25ea8c0a19ff6597ee1da25d3a012c069bec703ff4fc307a2",
				FromIndex:   uint32(1),
				FromAmount:  int64(40000),
				//0.00100000
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
				ToAddr:   "328AhL6KRjEh4334WHcV8xuPWD1kTihJUo",
				ToAmount: int64(100000),
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
