package rylink

import (
	"encoding/hex"
	"fmt"
	"github.com/group-coldwallet/ltcserver/model/bo"
	"testing"
)

func TestLtcSignTxTpl(t *testing.T) {

	tpl := &bo.LtcTxTpl{
		TxIns: []bo.LtcTxInTpl{
			bo.LtcTxInTpl{
				FromAddr:    "LV7xSQkmk7eWmZL8uF5vVXrNFPr25QF66m",
				FromPrivkey: "TA37sxKuBmrs6Mn9ZAzdkCaXCPQms1LWRTEAjUw1Uf8PqZ8Y3dKY",
				FromTxid:    "ac3cdada71adc05a646d4b5a3ddab0bec865c9156ebae99d1f93a01ecd3dc9f6",
				FromIndex:   uint32(0),
				FromAmount:  int64(917450),
			},
		},
		TxOuts: []bo.LtcTxOutTpl{
			bo.LtcTxOutTpl{
				ToAddr:   "LV7xSQkmk7eWmZL8uF5vVXrNFPr25QF66m",
				ToAmount: int64(915450),
			},
			bo.LtcTxOutTpl{
				ToAddr:   "3NQmZbkRMhNnQDy6m4gyFKfQDsk2RmGavP",
				ToAmount: int64(1000),
			},
		},
	}
	fmt.Println(LtcSignTxTpl(tpl))

}

func TestPk(t *testing.T) {
	_, tplPkScriptByte, _ := CreatePayScript("LfVpv9ezCY8vLYsQ9LTGAx5hNBnw2P6gpe")
	t.Log(hex.EncodeToString(tplPkScriptByte))
	_, tplPkScriptByte, _ = CreatePayScript("MKvk94ctEjtQUz3osUups9ucwnepZZLbdT")
	t.Log(hex.EncodeToString(tplPkScriptByte))
}
