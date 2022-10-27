
import (
	"btmSign/common/validator"
	"btmSign/utils/keystore"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func BuildTx(params *validator.SignParams) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.TxOuts {
		if _, err := DecodeAddress(out.ToAddr); err != nil {
			return nil, err
		}
		outaddr, err := btcutil.DecodeAddress(out.ToAddr, NetParams)
		if err != nil {
			return nil, err
		}
		pubkeyscript, err := txscript.PayToAddrScript(outaddr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.ToAmountInt64, pubkeyscript))
		outMount += out.ToAmountInt64
	}
	for _, in := range params.TxIns {
		if _, err := DecodeAddress(in.FromAddr); err != nil {
			return nil, err
		}
		//txhash, err := wire.NewShaHashFromStr(in.FromTxid)
		txhash, err := chainhash.NewHashFromStr(in.FromTxid)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(txhash, in.FromIndex)
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
		inMount += in.FromAmountInt64
	}
	//额度是否足够
	if inMount < outMount {
		return nil, errors.New("insufficient mount")
	}
	//0.1 dash fee
	if inMount > outMount+100000 {
		return nil, errors.New("too many tx.fee")
	}
	return tx, nil
}
func GetPrivate(mchName, address string) (private []byte, err error) {
	if tmpA, err := keystore.KeystoreGetKeyA(mchName, address); err != nil {
		return nil, fmt.Errorf("doesn't find keyA for mch : %s , address : %s", mchName, address)
	} else if akey, err := keystore.Base64Decode([]byte(tmpA)); err != nil {
		return nil, fmt.Errorf("keyA base64 decode err:%v", err)
	} else if bkey, err := keystore.KeystoreGetKeyB(mchName, address); err != nil {
		return nil, fmt.Errorf("doesn't find keyB for mch : %s , address : %s", mchName, address)
	} else if privkey, err := keystore.AesCryptCfb([]byte(akey), []byte(bkey), false); err != nil {
		return nil, fmt.Errorf("aes crypt cfb failed : %s , address : %s", mchName, address)
	} else {
		return privkey, nil
	}

}
