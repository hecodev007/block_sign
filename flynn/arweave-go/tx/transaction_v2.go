package tx

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	arweave_go "github.com/JFJun/arweave-go"
	"github.com/JFJun/arweave-go/utils"
	"math/big"
	"reflect"
)

// NewTransactionV2 creates a brand new TransactionV2 struct
func NewTransactionV2(lastTx string, owner *big.Int, quantity string, target string, data []byte, reward string) *TransactionV2 {
	return &TransactionV2{
		lastTx:   lastTx,
		owner:    owner,
		quantity: quantity,
		target:   target,
		data:     data,
		reward:   reward,
		tags:     make([]Tag, 0),
		dataRoot: "",
		dataSize: fmt.Sprintf("%d", len(data)),
		dataTree: make([]interface{}, 0),
		format:   2,
	}
}

// Data returns the data of the TransactionV2
func (t *TransactionV2) Data() string {
	return utils.EncodeToBase64(t.data)
}

// Data returns the data of the TransactionV2
func (t *TransactionV2) Format() int {
	return t.format
}

// RawData returns the unencoded data
func (t *TransactionV2) RawData() []byte {
	return t.data
}

// LastTx returns the last TransactionV2 of the account
func (t *TransactionV2) LastTx() string {
	return t.lastTx
}

// Owner returns the Owner of the TransactionV2
func (t *TransactionV2) Owner() string {
	return utils.EncodeToBase64(t.owner.Bytes())
}

// Quantity returns the quantity of the TransactionV2
func (t *TransactionV2) Quantity() string {
	return t.quantity
}

// Reward returns the reward of the TransactionV2
func (t *TransactionV2) Reward() string {
	return t.reward
}

// Target returns the target of the TransactionV2
func (t *TransactionV2) Target() string {
	return t.target
}

// ID returns the id of the TransactionV2 which is the SHA256 of the signature
func (t *TransactionV2) ID() []byte {
	return t.id
}

// write by jun
func (t *TransactionV2) Txid() string {
	if len(t.id) <= 0 {
		return ""
	}
	return utils.EncodeToBase64(t.id)
}

// Hash returns the base64 RawURLEncoding of the TransactionV2 hash
func (t *TransactionV2) Hash() string {
	return utils.EncodeToBase64(t.id)
}

// Tags returns the tags of the TransactionV2 in plain text
func (t *TransactionV2) Tags() ([]Tag, error) {
	tags := []Tag{}
	for _, tag := range t.tags {
		// access name
		tagName, err := utils.DecodeString(tag.Name)
		if err != nil {
			return nil, err
		}
		tagValue, err := utils.DecodeString(tag.Value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, Tag{Name: string(tagName), Value: string(tagValue)})
	}
	return tags, nil
}

// RawTags returns the unencoded tags of the TransactionV2
func (t *TransactionV2) RawTags() []Tag {
	return t.tags
}

// AddTag adds a new tag to the TransactionV2
func (t *TransactionV2) AddTag(name string, value string) error {
	tag := Tag{
		Name:  utils.EncodeToBase64([]byte(name)),
		Value: utils.EncodeToBase64([]byte(value)),
	}
	t.tags = append(t.tags, tag)
	return nil
}

func (t *TransactionV2) SetID(id []byte) {
	t.id = id
}

// Signature returns the signature of the TransactionV2
func (t *TransactionV2) Signature() string {
	return utils.EncodeToBase64(t.signature)
}

// Sign creates the signing message, and signs it using the private key,
// It takes the SHA256 of the resulting signature to calculate the id of
// the signature
func (t *TransactionV2) Sign(w arweave_go.WalletSigner) (*TransactionV2, error) {
	// format the message
	msg, err := t.formatMsgBytes()
	//fmt.Println(msg)
	payload := t.deepHash(msg)
	//fmt.Println(payload)
	//fmt.Println(len(payload))
	data := arweaveHash(payload, SHA256)

	sig, err := w.Sign(data[:])
	//fmt.Println(sig)
	if err != nil {
		return nil, err
	}

	err = w.Verify(data[:], sig)
	if err != nil {
		return nil, err
	}

	id := sha256.Sum256(sig)

	idB := make([]byte, len(id))
	copy(idB, id[:])
	t.SetID(idB)

	t.signature = sig
	return t, nil
}

// MarshalJSON marshals as JSON
func (t *TransactionV2) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.formatV2())
}

// UnmarshalJSON unmarshals as JSON
func (t *TransactionV2) UnmarshalJSON(input []byte) error {
	txn := transactionV2JSON{}
	err := json.Unmarshal(input, &txn)
	if err != nil {
		return err
	}
	id, err := utils.DecodeString(txn.ID)
	if err != nil {
		return err
	}
	t.id = id
	t.format = txn.Format
	t.lastTx = txn.LastTx

	// gives me byte representation of the big num
	owner, err := utils.DecodeString(txn.Owner)
	if err != nil {
		return err
	}
	n := new(big.Int)
	t.owner = n.SetBytes(owner)

	t.tags = txn.Tags
	t.target = txn.Target
	t.quantity = txn.Quantity

	data, err := utils.DecodeString(txn.Data)
	if err != nil {
		return err
	}
	t.data = data
	t.reward = txn.Reward
	t.dataTree = txn.DataTree
	t.dataSize = txn.DataSize
	t.dataRoot = txn.DataRoot

	sig, err := utils.DecodeString(txn.Signature)
	if err != nil {
		return err
	}
	t.signature = sig

	return nil
}

func (t TransactionV2) formatMsg(acc, data []byte) []byte {
	ft := t.deepHashBytes(data)
	acc = append(acc, ft...)
	return arweaveHash(acc, Sha384)
}

// Format formats the TransactionV2 to a JSONTransactionV2 that can be sent out to an arweave node
func (t *TransactionV2) formatV2() *transactionV2JSON {
	return &transactionV2JSON{
		Format:    t.format,
		ID:        utils.EncodeToBase64(t.id),
		LastTx:    t.lastTx,
		Owner:     utils.EncodeToBase64(t.owner.Bytes()),
		Tags:      t.tags,
		Target:    t.target,
		Quantity:  t.quantity,
		Data:      utils.EncodeToBase64(t.data),
		Reward:    t.reward,
		Signature: utils.EncodeToBase64(t.signature),
		DataRoot:  t.dataRoot,
		DataSize:  t.dataSize,
		DataTree:  t.dataTree,
	}
}

func (t *TransactionV2) deepHashChunks(chunks []interface{}, acc []byte) []byte {
	if len(chunks) == 0 {
		return acc
	}
	var newAcc []byte
	dh := t.deepHash(chunks[0])
	tmpAcc := acc
	tmpAcc = append(tmpAcc, dh...)
	newAcc = arweaveHash(tmpAcc, Sha384)
	if len(chunks) == 1 {
		return newAcc
	}
	newChuck := chunks[1:]
	return t.deepHashChunks(newChuck, newAcc)

}

//
func (t *TransactionV2) deepHash(data interface{}) []byte {
	d := reflect.ValueOf(data)
	switch d.Kind() {
	case reflect.Slice:
		len := d.Len()
		acc := t.deepHashArray(len)
		newData := data.([]interface{})
		return t.deepHashChunks(newData, acc)
	case reflect.Ptr:
		bi := data.(*big.Int)
		return t.deepHashBigInt(bi)
	default:
		return nil
	}

}

func (t *TransactionV2) encodeTagData() (tags []interface{}) {
	if t.tags == nil || len(t.tags) == 0 {
		return tags
	}
	for _, tag := range t.tags {
		var tmp []interface{}
		name, _ := utils.DecodeString(tag.Name)
		nBytes := new(big.Int).SetBytes(name)
		tmp = append(tmp, nBytes)
		value, _ := utils.DecodeString(tag.Value)
		vBytes := new(big.Int).SetBytes(value)
		tmp = append(tmp, vBytes)

		tags = append(tags, tmp)
	}
	return tags
}

func (t *TransactionV2) formatMsgBytes() ([]interface{}, error) {
	format := new(big.Int).SetBytes(utils.StringToBuffer(fmt.Sprintf("%d", t.format)))
	lastTxBytes, err := utils.DecodeString(t.LastTx())
	if err != nil {
		return nil, err
	}
	lastTx := new(big.Int).SetBytes(lastTxBytes)
	targetBytes, err1 := utils.DecodeString(t.Target())
	if err1 != nil {
		return nil, err1
	}
	target := new(big.Int).SetBytes(targetBytes)
	data_size := utils.StringToBuffer(t.dataSize)
	ds := new(big.Int).SetBytes(data_size)
	quantity := utils.StringToBuffer(t.quantity)
	quan := new(big.Int).SetBytes(quantity)
	reward := utils.StringToBuffer(t.reward)
	rw := new(big.Int).SetBytes(reward)
	data_root_bytes, err2 := utils.DecodeString(t.dataRoot)
	if err2 != nil {
		return nil, err
	}
	data_root := new(big.Int).SetBytes(data_root_bytes)
	tags := t.encodeTagData()

	msg := []interface{}{
		format,
		t.owner,
		target,
		quan,
		rw,
		lastTx,
		tags,
		ds,
		data_root,
	}
	return msg, nil
}

const (
	SHA256 = iota
	Sha384
)

func arweaveHash(msg []byte, alg int) []byte {
	var data []byte
	switch alg {
	case 0:
		d := sha256.Sum256(msg)
		data = d[:]
	case 1:
		h := sha512.New384()
		h.Write(msg)
		d := h.Sum(nil)
		data = d[:]
	}
	return data
}

func (t *TransactionV2) deepHashArray(length int) []byte {
	var tag []byte
	list := []byte("list")

	len := []byte(fmt.Sprintf("%d", length))
	tag = append(tag, list...)
	tag = append(tag, len...)

	return arweaveHash(tag, Sha384)
}
func (t *TransactionV2) deepHashBigInt(bi *big.Int) []byte {
	data := bi.Bytes()
	return t.deepHashBytes(data)
}

func (t *TransactionV2) deepHashBytes(data []byte) []byte {
	var (
		tag    []byte
		result []byte
	)
	list := []byte("blob")
	len := []byte(fmt.Sprintf("%d", len(data)))
	tag = append(tag, list...)
	tag = append(tag, len...)
	h1 := arweaveHash(tag, Sha384)
	h2 := arweaveHash(data, Sha384)
	result = append(result, h1...)
	result = append(result, h2...)
	return arweaveHash(result, Sha384)
}
