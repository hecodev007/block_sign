package crust

import (
	"encoding/hex"
	"errors"
	"fmt"
	sr25519 "github.com/ChainSafe/go-schnorrkel"
	"github.com/ChainSafe/gossamer/lib/runtime/extrinsic"
	"github.com/JFJun/substrate-go/ss58"
	r255 "github.com/gtank/ristretto255"
)

func GenerateKey() ([]byte, []byte, error) {
	secret, err := sr25519.GenerateMiniSecretKey()
	if err != nil {
		return nil, nil, err
	}
	if len(secret.Encode()) != 32 {
		return nil, nil, errors.New("private key or public key length i not equal 32")
	}
	priv := secret.Encode()
	pub := secret.Public().Encode()
	return pub[:], priv[:], nil
}
func PrivateToAddress(pri string, prifix []byte) (address string, err error) {
	//0x16a3fccddbaf51e5688d8d8dad0d0464e14567b251263dad0551433b54c84bea
	priBytes, err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	var pri32 [32]byte
	copy(pri32[:], priBytes)
	private, err := sr25519.NewMiniSecretKeyFromRaw(pri32)
	if err != nil {
		return "", err
	}
	pub := private.Public().Encode()
	return PubKeyToAddress(pub[:], prifix)

}
func CreateAddress(prefix []byte) (address string, private string, err error) {
	pub, pri, err := GenerateKey()
	if err != nil {
		return "", "", err
	}
	address, err = PubKeyToAddress(pub, prefix)
	return address, hex.EncodeToString(pri), err
}
func PubKeyToAddress(pubKey, prefix []byte) (string, error) {
	return ss58.Encode(pubKey, prefix)
}
func Sign(from, to string, amount, nonce uint64, pri string) (string, error) {
	_from, err := ss58.DecodeToPub(from)
	if err != nil {
		return "", errors.New("error from address:" + from)
	}
	_to, err := ss58.DecodeToPub(to)
	if err != nil {
		return "", errors.New("error to address:" + to)
	}
	var fromBytes32, toBytes32 [32]byte
	copy(fromBytes32[0:32], _from)
	copy(toBytes32[0:32], _to)
	tx := extrinsic.NewTransfer(fromBytes32, toBytes32, amount, nonce)
	pribytes, err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	//private, err := sr255192.NewPrivateKey(pribytes)
	//if err != nil {
	//	return "", err
	//}
	message, _ := tx.Encode()
	if err != nil {
		return "", err
	}
	signed, err := sign(pribytes, message)
	if err != nil {
		return "", err
	}
	rawTx := append(message, signed...)
	return hex.EncodeToString(rawTx), nil

}
func SignTx(from, to string, amount, nonce, fee uint64, pri string, genesisHash, blockHash string, blockNumber uint64, specVersion, transactionVersion uint32, callId string) (string, error) {
	//(from, to string, amount, nonce, fee uint64, genesisHash, blockHash string, blockNumber uint64, specVersion, transactionVersion uint32, callId string)
	tx := CreateTransaction(from, to, amount, nonce, fee, genesisHash, blockHash, blockNumber, specVersion, transactionVersion, callId)
	return SignTransaction(tx, pri)
}
func sign(privateKey, message []byte) ([]byte, error) {
	var sigBytes []byte
	var key, nonce [32]byte
	copy(key[:], privateKey[:32])
	signContext := sr25519.NewSigningContext([]byte("substrate"), message)
	if len(privateKey) == 32 { // Is seed

		sk, err := sr25519.NewMiniSecretKeyFromRaw(key)
		if err != nil {
			return nil, err
		}

		signContext.AppendMessage([]byte("proto-name"), []byte("Schnorr-sig"))
		pub := sk.Public()
		pubc := pub.Compress()
		signContext.AppendMessage([]byte("sign:pk"), pubc[:])

		r, err := sr25519.NewRandomScalar()
		if err != nil {

			return nil, err
		}
		R := r255.NewElement().ScalarBaseMult(r)
		signContext.AppendMessage([]byte("sign:R"), R.Encode([]byte{}))

		// form k
		kb := signContext.ExtractBytes([]byte("sign:c"), 64)
		k := r255.NewScalar()
		k.FromUniformBytes(kb)

		// form scalar from secret key x
		x, err := sr25519.ScalarFromBytes(sk.ExpandEd25519().Encode())
		if err != nil {
			return nil, err
		}
		// s = kx + r
		s := x.Multiply(x, k).Add(x, r)
		sig := sr25519.Signature{R: R, S: s}
		sbs := sig.Encode()
		sigBytes = sbs[:]
		varifySigContent := sr25519.NewSigningContext([]byte("substrate"), message)
		if !sk.Public().Verify(&sig, varifySigContent) {
			return nil, errors.New("verify sign error")
		}
	} else if len(privateKey) == 64 { //Is private key
		copy(nonce[:], privateKey[32:])
		sk := sr25519.NewSecretKey(key, nonce)
		sig, err := sk.Sign(signContext)
		if err != nil {
			return nil, fmt.Errorf("sr25519 sign error,err=%v", err)
		}
		sbs := sig.Encode()
		sigBytes = sbs[:]
		pub, _ := sk.Public()
		if !pub.Verify(sig, sr25519.NewSigningContext([]byte("substrate"), message)) {
			return nil, errors.New("verify sign error")
		}
	}
	return sigBytes[:], nil

}
