package avax

import (
	"avaxDataServer/common/log"
	"encoding/json"
	"strconv"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/codec"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

type Tx struct {
	UnsignedTx  UnsignedTx  `json:"unsignedTx"`
	Credentials interface{} `json:"credentials"`
}
type UnsignedTx struct {
	NetworkID    int64    `json:"networkID"`
	BlockchainID string   `json:"blockchainID"`
	Outputs      []output `json:"outputs"`
	Inputs       []input  `json:"inputs"`
	Memo         string   `json:"memo"`
}
type output struct {
	AssetID string `json:"assetID"`
	Output  struct {
		Amount    int64    `json:"amount"`
		Locktime  uint64   `json:"locktime"`
		Threshold int64    `json:"threshold"`
		Addresses []string `json:"addresses"`
	}
}
type input struct {
	TxID        string `json:"txID"`
	OutputIndex int    `json:"outputIndex"`
	AssetID     string `json:"assetID"`
	Input       struct {
		Amount           int64    `json:"amount"`
		Addresses        []string `json:"addresses"`
		SignatureIndices []interface{}
	}
}

func ToTransaction(rawtx *Tx, txid string) (tx Transaction) {
	tx.ID = txid
	tx.ChainID = rawtx.UnsignedTx.BlockchainID
	tx.Timestamp = time.Now()
	for k, out := range rawtx.UnsignedTx.Outputs {
		tmpoutput := new(Output)
		tmpoutput.OutputIndex = uint64(k)
		for _, addr := range out.Output.Addresses {
			tmpoutput.Addresses = append(tmpoutput.Addresses, Address(addr))
		}
		tmpoutput.Amount = strconv.FormatInt(out.Output.Amount, 10)
		tmpoutput.Locktime = out.Output.Locktime
		tmpoutput.AssetID = out.AssetID
		tmpoutput.CreatedAt = time.Now()
		tx.Outputs = append(tx.Outputs, tmpoutput)
	}
	for _, input := range rawtx.UnsignedTx.Inputs {
		tmpInput := &Input{
			Output: new(Output),
		}
		tmpInput.Output.OutputIndex = uint64(input.OutputIndex)
		for _, addr := range input.Input.Addresses {
			tmpInput.Output.Addresses = append(tmpInput.Output.Addresses, Address(addr))
		}
		tmpInput.Output.Amount = strconv.FormatInt(input.Input.Amount, 10)
		tmpInput.Output.AssetID = input.AssetID
		tmpInput.Output.TransactionID = input.TxID
		tx.Inputs = append(tx.Inputs, tmpInput)
	}
	return
}

func ParseRawTransaction(rawTx string) (*Tx, error) {
	avmtx, err := ParseTx(rawTx)
	if err != nil {
		return nil, err
	}
	txjson, _ := json.Marshal(avmtx)
	tx := new(Tx)
	err = json.Unmarshal(txjson, tx)
	for k, out := range tx.UnsignedTx.Outputs {
		for index, addr := range out.Output.Addresses {
			tx.UnsignedTx.Outputs[k].Output.Addresses[index], err = StrtoAddr(addr)
			if err != nil {
				log.Warn(err.Error())
			}
		}
	}
	return tx, nil
}

func StrtoAddr(shot string) (string, error) {
	sid, err := ids.ShortFromString(shot)
	if err != nil {
		return "", err
	}
	return ShoToAddr(sid)
}
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
func ShoToAddr(id ids.ShortID) (address string, err error) {
	address, err = formatting.FormatBech32(constants.NetworkIDToHRP[1], id.Bytes())
	address = "X-" + address
	return
}
