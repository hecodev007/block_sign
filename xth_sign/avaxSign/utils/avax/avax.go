package avax

import (
	"avaxSign/common/conf"
	"avaxSign/utils/keystore"
	"fmt"
	"github.com/ava-labs/gecko/utils/codec"
	"github.com/ava-labs/gecko/vms/avm"
	"github.com/ava-labs/gecko/vms/components/avax"
	"github.com/ava-labs/gecko/vms/secp256k1fx"
	"log"

	//dagwallet "github.com/ava-labs/avash/wallets/dags"
	"errors"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/constants"
	"github.com/ava-labs/gecko/utils/crypto"
	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/ava-labs/gecko/utils/logging"
)

//mainnet 1,testnet 4
var networkId = uint32(1)

var chainId = ids.ID{}

func init() {
	var err error
	chainId, err = ids.FromString(conf.Cfg.Node.ChainID)
	if err != nil {
		panic(err.Error())
	}
	networkId = conf.Cfg.Node.NetworkID
}
func GenAccount() (address string, privkey string, err error) {
	factory := crypto.FactorySECP256K1R{}
	skGen, err := factory.NewPrivateKey()
	if err != nil {
		return "", "", err
	}

	sk := skGen.(*crypto.PrivateKeySECP256K1R)
	fb := formatting.CB58{}
	fb.Bytes = sk.Bytes()
	privkey = fb.String()
	//fmt.Printf("%x\n", sk.PublicKey().Address().Bytes())
	address, err = formatting.FormatBech32(constants.NetworkIDToHRP[networkId], sk.PublicKey().Address().Bytes())
	return "X-" + address, privkey, err
}
func AddressToShot(address string) (*ids.ShortID, error) {
	chainID, hrp, addr, err := formatting.ParseAddress(address)
	if err != nil {
		return nil, err
	}
	_ = chainID
	_ = hrp
	shotid, err := ids.ToShortID(addr)
	return &shotid, err
}
func ShoToAddr(id ids.ShortID) (address string, err error) {
	address, err = formatting.FormatBech32(constants.NetworkIDToHRP[networkId], id.Bytes())
	address = "X-" + address
	return
}
func NewWallet(privkey string, txFee uint64) (w *Wallet, err error) {

	//w = dagwallet.NewWallet(networkID, chainID, 0)
	w, err = NewAvaxWallet(logging.NoLog{}, networkId, chainId, txFee)
	if err != nil {
		return w, err
	}

	if privkey != "" {
		factory := crypto.FactorySECP256K1R{}
		fb := formatting.CB58{}
		if err = fb.FromString(privkey); err != nil {
			return nil, err
		}
		pk, err := factory.ToPrivateKey(fb.Bytes)
		if err != nil {
			return nil, err
		}

		w.ImportKey(pk.(*crypto.PrivateKeySECP256K1R))
	}
	return w, nil
}
func ImportKey(w *Wallet, private string) error {
	factory := crypto.FactorySECP256K1R{}
	fb := formatting.CB58{}
	if err := fb.FromString(private); err != nil {
		return err
	}
	pk, err := factory.ToPrivateKey(fb.Bytes)
	if err != nil {
		return err
	}

	w.ImportKey(pk.(*crypto.PrivateKeySECP256K1R))
	return nil
}
func AddUtxos(w *Wallet, utxos []string, mchName string) (err error) {
	for _, v := range utxos {
		fmt.Printf("add utxos")
		//fb.FromString(v)
		utxo, err := ParseUtxo(v)
		if err != nil {
			return err
		}
		switch out := utxo.Out.(type) {
		case *secp256k1fx.MintOutput:
		case *secp256k1fx.TransferOutput:
			if len(out.OutputOwners.Addrs) <= 0 {
				return errors.New("utxo empty owner")
			}
			address, err := ShoToAddr(out.OutputOwners.Addrs[0])
			log.Println(address)
			if err != nil {
				return nil
			}
			private, err := GetPrivate(mchName, address)
			if err != nil {
				return err
			}

			if err := ImportKey(w, string(private)); err != nil {
				log.Println(err.Error())
				//return err
			}

		default:
			return errors.New("err utxo type")
		}
		w.AddUTXO(utxo)
	}
	return nil
}
func GetPrivate(mchName, address string) (private []byte, err error) {
	//return []byte("RZJ9ky62feWgLLNjvynFbhtf1pgdpXkEUy68rpGCviMuNzevUipx"), nil
	//get mch akey
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

//解析avm.getTx,tx
func ParseTx(rawTx string) (*avm.Tx, error) {
	fb := formatting.CB58{}
	fb.FromString(rawTx)
	tx := new(avm.Tx)
	c := codec.NewDefault()
	{
		c.RegisterType(&avm.BaseTx{})
		c.RegisterType(&avm.CreateAssetTx{})
		c.RegisterType(&avm.OperationTx{})
		c.RegisterType(&avm.ImportTx{})
		c.RegisterType(&avm.ExportTx{})
		c.RegisterType(&secp256k1fx.TransferInput{})
		c.RegisterType(&secp256k1fx.MintOutput{})
		c.RegisterType(&secp256k1fx.TransferOutput{})
		c.RegisterType(&secp256k1fx.MintOperation{})
		c.RegisterType(&secp256k1fx.Credential{})
	}
	err := c.Unmarshal(fb.Bytes, &tx)
	return tx, err
}
func MarsonTx(tx *avm.Tx) (string, error) {
	//txjson, _ := json.Marshal(tx)
	//fmt.Println("TXJSON:", string(txjson))
	c := codec.NewDefault()
	{
		c.RegisterType(&avm.BaseTx{})
		c.RegisterType(&avm.CreateAssetTx{})
		c.RegisterType(&avm.OperationTx{})
		c.RegisterType(&avm.ImportTx{})
		c.RegisterType(&avm.ExportTx{})
		c.RegisterType(&secp256k1fx.TransferInput{})
		c.RegisterType(&secp256k1fx.MintOutput{})
		c.RegisterType(&secp256k1fx.TransferOutput{})
		c.RegisterType(&secp256k1fx.MintOperation{})
		c.RegisterType(&secp256k1fx.Credential{})
	}
	bytes, err := c.Marshal(tx)
	if err != nil {
		return "", err
	}
	fb := formatting.CB58{}
	fb.Bytes = bytes
	return fb.String(), nil
}

func ParseUtxo(rawUtxo string) (*avax.UTXO, error) {
	fb := formatting.CB58{}
	fb.FromString(rawUtxo)
	c := codec.NewDefault()
	{
		c.RegisterType(&avm.BaseTx{})
		c.RegisterType(&avm.CreateAssetTx{})
		c.RegisterType(&avm.OperationTx{})
		c.RegisterType(&avm.ImportTx{})
		c.RegisterType(&avm.ExportTx{})
		c.RegisterType(&secp256k1fx.TransferInput{})
		c.RegisterType(&secp256k1fx.MintOutput{})
		c.RegisterType(&secp256k1fx.TransferOutput{})
		c.RegisterType(&secp256k1fx.MintOperation{})
		c.RegisterType(&secp256k1fx.Credential{})
	}
	utxo := &avax.UTXO{}
	if err := c.Unmarshal(fb.Bytes, utxo); err != nil {
		return nil, err
	}
	//utxojson, _ := json.Marshal(utxo)
	//fmt.Println(string(utxojson))
	return utxo, nil
}
