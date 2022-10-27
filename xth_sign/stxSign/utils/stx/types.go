package stx

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/btcsuite/btcutil"

)

type Transaction struct {
	Version uint8 //mainnet,0x00;testnet 0x80
	Chainid uint32//const TESTNET_CHAIN_ID: u32 = 0x80000000;
					//const MAINNET_CHAIN_ID: u32 = 0x00000001;
	Auth SinglesigSpendingCondition
	//OnChainOnly = 1,  // must be included in a StacksBlock
	//OffChainOnly = 2, // must be included in a StacksMicroBlock
	//Any = 3,          // either
	AuthModel uint8
	//Allow = 0x01, // allow any other changes not specified
	//Deny = 0x02,  // deny any other changes not specified
	PostConditionMode  uint8
	//待定
	PostConditions uint32
	Payload TokenTransfer
}

func (tx *Transaction)Serialize()(data []byte){
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tx)
	return bytesBuffer.Bytes()
}
func (tx *Transaction) Txid()(txhash string){
	tx_fee := tx.Auth.TxFee
	tx.Auth.TxFee = 0
	nonce := tx.Auth.Nonce
	tx.Auth.Nonce = 0
	Sigature :=tx.Auth.Sigature
	tx.Auth.Sigature = [65]byte{}
	defer func() {
		tx.Auth.Nonce = nonce
		tx.Auth.TxFee = tx_fee

		tx.Auth.Sigature = Sigature
	}()

	txbytes := tx.Serialize()
	//log.Info(hex.EncodeToString(txbytes))
	//txbytes,_= hex.DecodeString("00000000000400630a2cc9dc3e5186f8ceacf77800ce9da332e67300000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030100000000000501ffffffffffffffffffffffffffffffffffffffff000000000000007b00000000000000000000000000000000000000000000000000000000000000000000")
	h :=sha512.New512_256()
	h.Write(txbytes)
	sum := h.Sum(nil)
	//log.Info("txid",hex.EncodeToString(sum))
	return hex.EncodeToString(sum)
}
func (tx *Transaction)Sign(priHex string)( error){
	return tx.Sign2(priHex)
}
func (tx *Transaction)Sign2(wifpri string) error{
	wif,err :=btcutil.DecodeWIF(wifpri)
	if err != nil{
		return err
	}
	tx.Auth.KeyEncoding = 0x00

	pri := wif.PrivKey.Serialize()
	tx_fee := tx.Auth.TxFee
	//tx.Auth.TxFee = 0
	nonce := tx.Auth.Nonce
	//tx.Auth.Nonce = 0

	txid := tx.Txid()
	//log.Info(txid)
	txhash,_ := hex.DecodeString(txid)
	hashbuff := bytes.NewBuffer([]byte{})
	binary.Write(hashbuff, binary.BigEndian, txhash)
	binary.Write(hashbuff, binary.BigEndian, uint8(0x04))
	binary.Write(hashbuff, binary.BigEndian, tx_fee)
	binary.Write(hashbuff, binary.BigEndian, nonce)
	h :=sha512.New512_256()
	h.Write(hashbuff.Bytes())
	signhash := h.Sum(nil)
	sig,err :=secp256k1.Sign(signhash,pri)
	if err != nil{
		return err
	}
	copy(tx.Auth.Sigature[0:1],sig[64:65])
	copy(tx.Auth.Sigature[1:],sig[:64])
	return nil
}
func (tx *Transaction)Sign1(priHex string)( error){
	pri,err := hex.DecodeString(priHex)
	if err != nil{
		return err
	}
	tx_fee := tx.Auth.TxFee
	//tx.Auth.TxFee = 0
	nonce := tx.Auth.Nonce
	//tx.Auth.Nonce = 0

	txid := tx.Txid()
	//log.Info(txid)
	txhash,_ := hex.DecodeString(txid)
	hashbuff := bytes.NewBuffer([]byte{})
	binary.Write(hashbuff, binary.BigEndian, txhash)
	binary.Write(hashbuff, binary.BigEndian, uint8(0x04))
	binary.Write(hashbuff, binary.BigEndian, tx_fee)
	binary.Write(hashbuff, binary.BigEndian, nonce)
	h :=sha512.New512_256()
	h.Write(hashbuff.Bytes())
	signhash := h.Sum(nil)
	//println("signhash",hex.EncodeToString(signhash))
	//0b6c70aefa232c3097a7196592a25fe60dd991bad094b8965cb039c8032a9986
	//872a8f8ec8291f84eadfa7134eb5103d60b3cb08f9e4df5b68b8e8569f11732b
	//872a8f8ec8291f84eadfa7134eb5103d60b3cb08f9e4df5b68b8e8569f11732b
	//prikey,_ :=btcec.PrivKeyFromBytes(cbtcec.S256(),pri)
	sig,err :=secp256k1.Sign(signhash,pri)

	//004e5f96b6ff51388624b98dbd2a2033908db3f0c5a2f8d9bfc8a59fdb6750e0d849198faf8309809b98de8b275fb10ff571f3588a5fb510b521cc02aa1841300d
	//  4e5f96b6ff51388624b98dbd2a2033908db3f0c5a2f8d9bfc8a59fdb6750e0d849198faf8309809b98de8b275fb10ff571f3588a5fb510b521cc02aa1841300d00
	//sig,err :=prikey.Sign(signhash)
	if err != nil{
		return err
	}
	copy(tx.Auth.Sigature[0:1],sig[64:65])
	copy(tx.Auth.Sigature[1:],sig[:64])
	//println("签名:",hex.EncodeToString(tx.Auth.Sigature[:]))
	//6f92e1757b63466cbd1898b5b0dcf58757e22f9ca360ce23ea15395d1ee13e38
	//

	return nil
}
type TokenTransfer struct {
	TransferType uint8 //0x00
	Receipient StandardPrincipalData
	Amount uint64
	Memo [34]byte
}
type StandardPrincipalData struct {
	Type uint8 //0x05 PrincipalStandard
	Version uint8 //0x01
	Bytes [20]byte
}
type SinglesigSpendingCondition struct {
	Type uint8 //TransactionAuthFlags 0x04;AuthSponsored 0x05;这里默认0x04
	HashModel uint8 // P2PKH = 0x00;P2WPKH = 0x02
	Signer [20]byte //
	Nonce uint64
	TxFee uint64
	KeyEncoding uint8 //Compressed = 0x00,	Uncompressed = 0x01,
	Sigature [65]byte
}
func (ssc *SinglesigSpendingCondition)Serialise()[]byte{
	return nil
}
type TokenTransferMemo [32]byte