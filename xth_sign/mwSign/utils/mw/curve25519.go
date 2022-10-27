package mw

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"

	gocurve25519 "github.com/moonfruit/go-curve25519"
)

func Sign(data []byte, pri string) (string, error) {
	private, err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	prikey := gocurve25519.NewPrivateKey(private)
	sign := prikey.Sign(data)
	unsigntx := hex.EncodeToString(data)
	signature := hex.EncodeToString(sign[:])
	return unsigntx[:192] + signature + unsigntx[320:], nil
	//return hex.EncodeToString(data) + hex.EncodeToString(sign[:]), nil
}

func NewTransaction() *Transaction {
	tx := new(Transaction)
	tx.Version = 3
	return tx
}

type Transaction struct {
	Type       uint8    `json:"type"`
	Version    uint8    `json:"version"` //uint4
	Subtype    uint8    `json:"subtype"` //uint4
	Timestamp  uint32   `json:"timestamp"`
	Deadline   uint16   `json:"deadline"`
	PublickKey [32]byte `json:"senderPublicKey"`
	Recipient  uint64
	AmountNQT  uint64
	FeeNQT     uint64 `json:"feeNQT"`
	FullHash   string `json:"fullHash"`
}

func (tx *Transaction) SetRecipient(toAccountId string) error {
	acc_id, err := strconv.ParseUint(toAccountId, 10, 64)
	if err != nil {
		return err
	}
	tx.Recipient = acc_id
	return nil
}
func (tx *Transaction) Seriallize() []byte {
	ret := make([]byte, 64, 64)
	ret[0] = tx.Type
	ret[1] = tx.Version<<4 + tx.Subtype
	binary.LittleEndian.PutUint32(ret[2:6], tx.Timestamp)
	binary.LittleEndian.PutUint16(ret[6:8], tx.Deadline)
	copy(ret[8:40], tx.PublickKey[:])
	binary.LittleEndian.PutUint64(ret[40:48], tx.Recipient)
	binary.BigEndian.PutUint64(ret[48:56], tx.AmountNQT)
	binary.BigEndian.PutUint64(ret[56:64], tx.FeeNQT)
	return ret
}
func (tx *Transaction) Unseriallize(rawtx []byte) bool {
	if len(rawtx) < 64 {
		return false
	}
	tx.Type = rawtx[0]
	tx.Version = rawtx[1] & 0xF0 >> 4
	tx.Subtype = rawtx[1] & 0x0F
	tx.Timestamp = binary.LittleEndian.Uint32(rawtx[2:6])
	tx.Deadline = binary.LittleEndian.Uint16(rawtx[6:8])
	copy(tx.PublickKey[:], rawtx[8:40])
	//copy(tx.Recipient[:], rawtx[40:48])
	tx.Recipient = binary.LittleEndian.Uint64(rawtx[40:48])
	tx.AmountNQT = binary.BigEndian.Uint64(rawtx[48:56])
	tx.FeeNQT = binary.BigEndian.Uint64(rawtx[56:64])
	return true
}
