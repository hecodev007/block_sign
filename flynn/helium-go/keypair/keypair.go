package keypair

import (
	"bytes"
	"encoding/hex"
	"errors"
	cre "github.com/JFJun/helium-go/crypto"
	"github.com/JFJun/helium-go/utils"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ed25519"
)

const (
	NISTP256Version = iota
	Ed25519Version
	WIFVersion = 0x80
)

type Keypair struct {
	curve      cre.Curves //选择非对称加密方式
	privateKey []byte
	version    int
}

func New(version int) *Keypair {
	c := cre.NewCurve(version)
	ks := new(Keypair)
	ks.curve = c
	ks.version = version
	return ks
}

func NewKeypairFromWIF(version int, wif string) *Keypair {
	kp := New(version)
	err := kp.decodeWIF(wif)
	if err != nil {
		return nil
	}
	return kp
}

func NewKeypairFromHex(version int, privHex string) *Keypair {
	kp := New(version)
	data, err := hex.DecodeString(privHex)
	if err != nil {
		return nil
	}
	kp.privateKey = data
	return kp
}
func (kp *Keypair) GenerateKey() ([]byte, []byte) {
	return kp.curve.GenerateKey()
}

/*
新生成wif格式的私钥以及对应的地址
*/
func (kp *Keypair) GenerateWifAndAddress() (wif, address string) {
	priv, pub := kp.GenerateKey()
	//kp.privateKey = priv
	if len(priv) != 32 || len(pub) != 32 {
		return "", ""
	}
	wif = kp.toWif(priv)
	if wif == "" {

		return wif, ""
	}
	address = kp.CreateAddress(pub)
	if address == "" {
		return "", address
	}
	return
}

func (kp *Keypair) toWif(priv []byte) string {
	var private []byte
	if priv != nil {
		private = priv
	} else {
		private = kp.privateKey
	}
	if len(private) != 32 {
		return ""
	}
	payload := append([]byte{WIFVersion}, private...)
	c := utils.DoubleSha256(payload)
	checkSum := c[:4]
	pc := append(payload, checkSum...)
	return base58.Encode(pc)
}

func (kp *Keypair) decodeWIF(wif string) error {
	pc := base58.Decode(wif)
	checkSum := pc[len(pc)-4:]
	payload := pc[:len(pc)-4]
	if len(payload) != 33 {
		return errors.New("private key len is not correct")
	}
	c := utils.DoubleSha256(payload)
	checkSum2 := c[:4]
	if !bytes.Equal(checkSum, checkSum2) {
		return errors.New("checkSum is not equal")
	}
	private := payload[1:]
	kp.privateKey = private
	return nil
}

func (kp *Keypair) CreateAddressable() *Addressable {
	if kp.privateKey == nil {
		return nil
	}
	priv := ed25519.NewKeyFromSeed(kp.privateKey)
	pub := make([]byte, 32)
	copy(pub, priv[32:])
	address := kp.CreateAddress(pub)
	var bin []byte
	v := kp.curve.GetVersion()
	bin = append(bin, v...)
	bin = append(bin, pub...)
	aa := new(Addressable)
	aa.base58 = address
	aa.bin = bin
	aa.publicKey = pub
	return aa
}

func (kp *Keypair) CreateAddress(publicKey []byte) string {
	var (
		payload  []byte
		vpayload []byte
	)
	v := kp.curve.GetVersion() //曲线版本号=》 1->ed25519 0-> NIST p256
	payload = append(v, publicKey[:]...)
	version := []byte{0} //主网版本号
	vpayload = append(version, payload...)
	//double sha256
	checksum := utils.DoubleSha256(vpayload)[:4]
	vpayload = append(vpayload, checksum...)

	return base58.Encode(vpayload)
}

func (kp *Keypair) SetPrivateKey(privateKey []byte) {
	kp.privateKey = privateKey
	return
}
func (kp *Keypair) SetPrivateKeyToNull() {
	if kp.privateKey != nil {
		tmp := make([]byte, len(kp.privateKey))
		kp.privateKey = tmp
		return
	}
}

func (kp *Keypair) Sign(message []byte) ([]byte, error) {
	if kp.privateKey == nil {
		return nil, errors.New("private key is null")
	}
	var data []byte
	if kp.version == 1 {
		privKey := ed25519.NewKeyFromSeed(kp.privateKey)
		data = ed25519.Sign(privKey, message)
	} else {
		//todo
		//privKey:=nist_p256.NewNISTP256PrivateBySeed(kp.privateKey)
		//ecdsa.Sign(rand.Reader,privKey,message)
	}
	return data, nil
}
