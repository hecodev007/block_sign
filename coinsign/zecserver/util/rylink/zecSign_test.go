package rylink

import (
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/group-coldwallet/zecserver/model/bo"
	"github.com/iqoption/zecutil"
	"github.com/shopspring/decimal"
	"testing"
)

//必须要使用这一方法
func TestCheckAddr(t *testing.T) {
	addr, err := zecutil.DecodeAddress("t1YfKL1iKd9YFPP5mWop3qywUegCczmPGik", "mainnet")
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(addr.EncodeAddress())
	}

	addr2, err2 := zecutil.DecodeAddress("t3ZMLFZCmB3rdUYnCduUeGcJQoCVLViCiTW", "mainnet")
	if err2 != nil {
		t.Error(err.Error())
	} else {
		t.Log(addr2.EncodeAddress())
	}
}

func TestCheckAddr2(t *testing.T) {
	decoded := base58.Decode("t1YfKL1iKd9YFPP5mWop3qywUegCczmPGik")
	decoded2 := base58.Decode("t3ZMLFZCmB3rdUYnCduUeGcJQoCVLViCiTW")
	addr, err := btcutil.NewAddressPubKeyHash(decoded[2:len(decoded)-4], netParams)
	if err != nil {
		t.Log("err")
		t.Log(err)
	}
	//decoded2 := base58.Decode("t3ZMLFZCmB3rdUYnCduUeGcJQoCVLViCiTW")
	addr2, err2 := btcutil.NewAddressPubKeyHash(decoded2[2:len(decoded)-4], netParams)
	if err2 != nil {
		t.Log("err")
		t.Log(err2)
	}
	t.Log(addr)
	t.Log(addr2)

	t.Log(addr.String() == addr2.String())

}

func TestSignTxTpl(t *testing.T) {
	tpl := &bo.ZecTxTpl{
		ExpiryHeight: 1095541,
		TxIns: []bo.ZecTxInTpl{
			bo.ZecTxInTpl{
				FromAddr:         "t1S5Uf22M9w2y8Ms3WjByJHiUFxaUB2Gq5M",
				FromPrivkey:      "L4bRWqFr4icTcUR5WSLYPBzCqBFUA24Mj2VHsLi4RRxnXb4oQWzE",
				FromTxid:         "c6555f7bb397e7408e0b5930c37ba8ef23f3ab7767b211980fa0f74b8940adc7",
				FromIndex:        0,
				FromAmount:       decimal.NewFromFloat(0.02).Shift(8).IntPart(),
				FromScriptPubKey: "",
				FromRedeemScript: "",
			},
		},
		TxOuts: []bo.ZecTxOutTpl{
			bo.ZecTxOutTpl{
				ToAddr:   "t1S5Uf22M9w2y8Ms3WjByJHiUFxaUB2Gq5M",
				ToAmount: decimal.NewFromFloat(0.01997).Shift(8).IntPart(),
			},
			bo.ZecTxOutTpl{
				ToAddr:   "t3YruMWrzqCaxmWNPs44caupjZ5hF73CJoR",
				ToAmount: decimal.NewFromFloat(0.00001).Shift(8).IntPart(),
			},
		},
	}
	t.Log(SignTxTpl(tpl))
	//t.Log(SignTxTplError(tpl))

}

func TestSignTxTpl2(t *testing.T) {
	tpl := &bo.ZecTxTpl{
		ExpiryHeight: 613116,
		TxIns: []bo.ZecTxInTpl{
			bo.ZecTxInTpl{
				FromAddr:         "t3X8aoRJz4YUamG99pcFFLzMZkt1XzuYZu2",
				FromPrivkey:      "L4jajt3y3qEkrwJh6fbzbnY8qp4d12F4e5Braw6fPy6WZmfB21DG",
				FromTxid:         "c5cebeb3f9907e9fe21d3ab869425da54ef2d96a2821c83f90e2e916f594b1be",
				FromIndex:        0,
				FromAmount:       100000,
				FromScriptPubKey: "",
				FromRedeemScript: "",
			},
		},
		TxOuts: []bo.ZecTxOutTpl{
			bo.ZecTxOutTpl{
				ToAddr:   "t3X8aoRJz4YUamG99pcFFLzMZkt1XzuYZu2",
				ToAmount: 9000,
			},
			bo.ZecTxOutTpl{
				ToAddr:   "t1QYUY3ZmxCca9ZM8VN8ABSVyqFJf4dwiSq",
				ToAmount: 90000,
			},
			//bo.ZecTxOutTpl{
			//	ToAddr:   "t1WSZsspYWeACg6SihWaevMzdcMipPARCS5",
			//	ToAmount: 8000,
			//},
		},
	}
	t.Log(SignTxTpl(tpl))

}

func TestSignTxTpl3(t *testing.T) {
	tpl := &bo.ZecTxTpl{
		ExpiryHeight: 808911,
		TxIns: []bo.ZecTxInTpl{
			bo.ZecTxInTpl{
				FromAddr:         "t1SM2aDhvz535zJ43zQnb393x3YLUf4sAmZ",
				FromPrivkey:      "KxzdJKNxv4cd3wrYoLNyaQTkie1Ymy7DXR4B6xYhu1xdnE4w3GFs",
				FromTxid:         "0dfd6e400ee46b197e8cada439b4cbc145ceff65db1bf798cc87d7ed348f9e2a",
				FromIndex:        0,
				FromAmount:       30000,
				FromScriptPubKey: "76a9145cf0bea581fee7556db3087422478848e0db814488ac",
				FromRedeemScript: "",
			},
		},
		TxOuts: []bo.ZecTxOutTpl{
			bo.ZecTxOutTpl{
				ToAddr:   "t1SM2aDhvz535zJ43zQnb393x3YLUf4sAmZ",
				ToAmount: 29000,
			},
		},
	}
	t.Log(SignTxTpl(tpl))

}

func TestSignTxTpl4(t *testing.T) {
	tpl := &bo.ZecTxTpl{
		ExpiryHeight: 610116,
		TxIns: []bo.ZecTxInTpl{
			bo.ZecTxInTpl{
				FromAddr:         "t3X8aoRJz4YUamG99pcFFLzMZkt1XzuYZu2",
				FromPrivkey:      "L4jajt3y3qEkrwJh6fbzbnY8qp4d12F4e5Braw6fPy6WZmfB21DG",
				FromTxid:         "c5cebeb3f9907e9fe21d3ab869425da54ef2d96a2821c83f90e2e916f594b1be",
				FromIndex:        1,
				FromAmount:       90000,
				FromScriptPubKey: "a91489dd7593820226541686faae141a5807f997787987",
				FromRedeemScript: "",
			},
			bo.ZecTxInTpl{
				FromAddr:         "t3X8aoRJz4YUamG99pcFFLzMZkt1XzuYZu2",
				FromPrivkey:      "L4jajt3y3qEkrwJh6fbzbnY8qp4d12F4e5Braw6fPy6WZmfB21DG",
				FromTxid:         "c5cebeb3f9907e9fe21d3ab869425da54ef2d96a2821c83f90e2e916f594b1be",
				FromIndex:        1,
				FromAmount:       8000,
				FromScriptPubKey: "a91489dd7593820226541686faae141a5807f997787987",
				FromRedeemScript: "",
			},
		},
		TxOuts: []bo.ZecTxOutTpl{
			bo.ZecTxOutTpl{
				ToAddr:   "t3ZMLFZCmB3rdUYnCduUeGcJQoCVLViCiTW",
				ToAmount: 1006000,
			},
		},
	}
	t.Log(SignTxTpl(tpl))

}

func TestCheckAddressType(t *testing.T) {
	//addr,pk,err  := GetPayScriptByAddr("t1WSZsspYWeACg6SihWaevMzdcMipPARCS5")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//t.Log(addr.EncodeAddress())
	//t.Log(CheckAddressType(addr))
	//
	//
	//_,a,_,_ := txscript.ExtractPkScriptAddrs(pk,&chaincfg.MainNetParams)
	//t.Log(a[0].String())

}

func TestDecodeBlock(t *testing.T) {

	//testTxBytes,_ := hex.DecodeString("0400008085202f8901fce0f115bba01d8c56a2bcde1474a4dcbd15bde1d0ba205db5d6b0a39e6aa52a010000006a473044022027f453a2e03b715568206873a01f8d794eb1c63f7cb19e565b1e7dadb111ee2902207f913afd819a938590ee106376329bffeb2216c0d3eee5b7777a8c17c76837ed0121035ed8cf6b86efeefe5bc244cf35ec81f08f0b5d642fa2c45fc0a67fd49b3d96b2ffffffff0288130000000000001976a914264d16d0676bea59a536cf02cd22dc36c8532b1288ac905f0100000000001976a91459ffdcb235fc2de2276348dea8c8ba91b2f5e8f188ac00000000444f09000000000000000000000000")

	//tx, err := NewTxFromBytes(testTxBytes)
	//if err != nil {
	//	t.Errorf("Serialize: %v", err)
	//}
	//dd,_ := json.Marshal(tx)
	//println(tx.Hash().String())
	//println(len(tx.MsgTx().TxOut))
	//println(string(dd))

}
