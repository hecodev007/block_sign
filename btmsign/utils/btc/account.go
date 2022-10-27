package btc

import (
	"btmSign/bytom/account"
	"btmSign/bytom/blockchain/pseudohsm"
	"btmSign/bytom/blockchain/signers"
	"btmSign/bytom/common"
	"btmSign/bytom/consensus"
	"btmSign/bytom/crypto"
	"btmSign/bytom/crypto/ed25519/chainkd"
	"btmSign/bytom/protocol/vm/vmutil"
	mnem "btmSign/bytom/wallet/mnemonic"
	"btmSign/common/conf"
	"btmSign/common/validator"
	"btmSign/net"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pborman/uuid"
	"math/rand"
	"strings"
	"time"
)

var NetParams *chaincfg.Params

func init() {
	NetParams = new(chaincfg.Params)
	NetParams.PubKeyHashAddrID = 0x00
	NetParams.ScriptHashAddrID = 0x05
	NetParams.PrivateKeyID = 0x80
	NetParams.HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
	NetParams.HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e}
}

//address:  bn1qwhe7tc54dp9ggfp09tx7z9spe22jqc3qthpw6s
//mnemonic:  certain novel muffin clarify voice exotic short useless ethics helmet rapid slim
//encode:  0x6365727461696e206e6f76656c206d756666696e20636c617269667920766f6963652065786f7469632073686f7274207573656c657373206574686963732068656c6d657420726170696420736c696d

func GenAccountBtm() (string, string, error) {
	_, mnemonic, err := createChainKDKey("", "", "en")
	if err != nil {
		return "", "", err
	}
	xpub, err := CreateKeyFromMnemonic("", "", *mnemonic)
	if err != nil {
		return "", "", err
	}
	xpubs := []chainkd.XPub{xpub.XPub}
	signer, err := signers.Create("account", xpubs, 1, uint64(1), 1)
	id := uuid.New()
	acc := &account.Account{Signer: signer, ID: id, Alias: strings.ToLower(strings.TrimSpace(""))}
	path, err := signers.Path(acc.Signer, signers.AccountKeySpace, false, uint64(1))
	if err != nil {
		return "", "", err
	}
	cp, err := createP2PKH(acc, path)
	if err != nil {
		return "", "", err
	}
	encode := hexutil.Encode([]byte(*mnemonic))
	//fmt.Println("address: ", cp.Address)
	//fmt.Println("mnemonic: ", *mnemonic)
	//fmt.Println("encode: ", encode)
	return cp.Address, encode, nil
}

func GetRandomPassword(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func GenAccountBtm2(alias string) (string, string, error) {
	password := GetRandomPassword(20)
	createKeyReq := net.CreateKeyRequest{
		Alias:    alias,
		Password: password,
		Language: "en",
	}
	url := conf.GetConfig().Node.Url
	createKeyResult, err := net.Post(url+net.CreateKey, createKeyReq)
	if err != nil {
		return "", "", err
	}
	var ckr net.CreateKeyResult
	err = json.Unmarshal([]byte(createKeyResult), &ckr)
	if err != nil {
		return "", "", err
	}
	if ckr.Status != "success" {
		return "", "", errors.New(createKeyResult)
	}
	xpub := ckr.Data.Xpub
	mnemonic := ckr.Data.Mnemonic

	//fmt.Println("Alias: ",alias)
	//fmt.Println("xpub: ",xpub)
	//fmt.Println("mnemonic: ",mnemonic)

	createAccountReq := net.CreateAccountRequest{
		Alias:     alias,
		RootXpubs: []string{xpub},
		Quorum:    1,
	}
	createAccountResult, err := net.Post(url+net.CreateAccount, createAccountReq)
	if err != nil {
		return "", "", err
	}
	var car net.CreateAccountResult
	err = json.Unmarshal([]byte(createAccountResult), &car)
	if err != nil {
		return "", "", err
	}

	if car.Status != "success" {
		return "", "", errors.New(createAccountResult)
	}

	id := car.Data.ID
	//fmt.Println("id: ",id)

	createAccountReceiverReq := net.CreateAccountReceiverRequest{
		AccountID: id,
	}

	createAccountReceiverResult, err := net.Post(url+net.CreateAccountReceiver, createAccountReceiverReq)
	if err != nil {
		return "", "", err
	}
	var carr net.CreateAccountReceiverResult
	err = json.Unmarshal([]byte(createAccountReceiverResult), &carr)
	if err != nil {
		return "", "", err
	}
	if carr.Status != "success" {
		return "", "", errors.New(createAccountReceiverResult)
	}

	address := carr.Data.Address
	pri := fmt.Sprintf("%s#%s#%s", id, password, mnemonic)
	bPri := hexutil.Encode([]byte(pri))

	//fmt.Println("encoderesult: ", bPri)
	//
	//fmt.Println("----------")
	//decode, _ := hexutil.Decode(bPri)
	//fmt.Println("decoderesult: ",string(decode))

	return address, bPri, err
}

func CreateKeyFromMnemonic(alias string, auth string, mnemonic string) (*pseudohsm.XPub, error) {
	seed := mnem.NewSeed(mnemonic, auth)
	_, xpub, err := chainkd.NewXKeys(bytes.NewBuffer(seed))
	if err != nil {
		return nil, err
	}
	return &pseudohsm.XPub{XPub: xpub, Alias: alias}, nil
}

func createP2PKH(accounts *account.Account, path [][]byte) (*account.CtrlProgram, error) {
	derivedXPubs := chainkd.DeriveXPubs(accounts.XPubs, path)
	derivedPK := derivedXPubs[0].PublicKey()
	pubHash := crypto.Ripemd160(derivedPK)

	address, err := common.NewAddressWitnessPubKeyHash(pubHash, &consensus.ActiveNetParams)
	if err != nil {
		return nil, err
	}

	control, err := vmutil.P2WPKHProgram([]byte(pubHash))
	if err != nil {
		return nil, err
	}

	return &account.CtrlProgram{
		AccountID:      accounts.ID,
		Address:        address.EncodeAddress(),
		ControlProgram: control,
	}, nil
}

func XpubToAddress(xpub chainkd.XPub) string {
	pub := xpub.PublicKey()
	pubHash := crypto.Ripemd160(pub)
	address, err := common.NewAddressWitnessPubKeyHash(pubHash, &consensus.ActiveNetParams) //主网
	if err != nil {
		fmt.Errorf("create address error,please try again")
		panic(err)
	}
	return address.EncodeAddress()
}

func createChainKDKey(alias string, auth string, language string) (*pseudohsm.XPub, *string, error) {
	//Generate a mnemonic for memorization or user-friendly seeds
	entropy, err := mnem.NewEntropy(128)
	if err != nil {
		return nil, nil, err
	}
	mnemonic, err := mnem.NewMnemonic(entropy, language)
	if err != nil {
		return nil, nil, err
	}
	xpub, err := createKeyFromMnemonic(alias, auth, mnemonic)
	if err != nil {
		return nil, nil, err
	}
	return xpub, &mnemonic, nil
}

func createKeyFromMnemonic(alias string, auth string, mnemonic string) (*pseudohsm.XPub, error) {
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := mnem.NewSeed(mnemonic, "")
	_, xpub, err := chainkd.NewXKeys(bytes.NewBuffer(seed))
	if err != nil {
		return nil, err
	}
	//id := uuid.NewRandom()
	//key := &pseudohsm.XKey{
	//	ID:      id,
	//	KeyType: "bytom_kd",
	//	XPub:    xpub,
	//	XPrv:    xprv,
	//	Alias:   alias,
	//}
	return &pseudohsm.XPub{XPub: xpub, Alias: alias, File: ""}, nil
}

func GenAccount() (address string, private string, err error) {
	pri, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", err
	}
	wif, err := btcutil.NewWIF(pri, NetParams, true)
	if err != nil {
		return "", "", err
	}
	pk := wif.SerializePubKey()
	pkhash, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(pk), NetParams)
	if err != nil {
		return "", "", err
	}
	address = pkhash.EncodeAddress()
	return address, wif.String(), nil
}

func BuildTx(params *validator.SignParams) (tx *wire.MsgTx, err error) {
	tx = wire.NewMsgTx(1)
	var outMount, inMount int64
	for _, out := range params.Outs {
		if _, err := btcutil.DecodeAddress(out.ToAddr, NetParams); err != nil {
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
	for _, in := range params.Ins {
		if _, err := btcutil.DecodeAddress(in.FromAddr, NetParams); err != nil {
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
	if inMount < outMount+100000 {
		return nil, errors.New("insufficient mount or fee(0.001)")
	}
	//max 1 fee
	if inMount > outMount+100000000 {
		return nil, errors.New("too many tx.fee")
	}
	return tx, nil
}
