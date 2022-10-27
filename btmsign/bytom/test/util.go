package test

import (
	"btmSign/bytom/common"
	"btmSign/bytom/protocol/vm/vmutil"
	"btmSign/common/validator"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"

	"btmSign/bytom/account"
	"btmSign/bytom/blockchain/pseudohsm"
	"btmSign/bytom/blockchain/txbuilder"
	"btmSign/bytom/consensus"
	"btmSign/bytom/crypto/ed25519/chainkd"
	"btmSign/bytom/database"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/event"
	"btmSign/bytom/protocol"
	"btmSign/bytom/protocol/bc"
	"btmSign/bytom/protocol/bc/types"
)

const (
	vmVersion    = 1
	assetVersion = 1
)

// MockChain mock chain with genesis block
func MockChain(testDB dbm.DB) (*protocol.Chain, *database.Store, *protocol.TxPool, error) {
	store := database.NewStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)
	chain, err := protocol.NewChain(store, txPool, dispatcher)
	return chain, store, txPool, err
}

func MakeChain(testDB dbm.DB) (*protocol.Chain, *database.Store, *protocol.TxPool, error) {
	store := database.NewStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)
	chain, err := protocol.NewChain(store, txPool, dispatcher)
	return chain, store, txPool, err
}

// MockUTXO mock a utxo
func MockUTXO(controlProg *account.CtrlProgram) *account.UTXO {
	hash, err := MustDecodeHash("45bc6f8b8ddcb3753cbcc97215891ec42c145e575cd3541cffa72d985d8b8b9d")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo := &account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = hash
	utxo.AssetID = *consensus.BTMAssetID
	utxo.Amount = 100000000
	utxo.SourcePos = 0
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	utxo.Change = controlProg.Change
	return utxo
}

func MockUTXO2(controlProg *account.CtrlProgram) []account.UTXO {
	var u []account.UTXO
	hash, err := MustDecodeHash("d25a673df4dbadcbb9edb0a46a98bcc57bbff6bf2d0b0ac646f44ebb5474b201")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo := account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = hash
	utxo.AssetID = *consensus.BTMAssetID
	utxo.Amount = 100000000
	utxo.SourcePos = 0
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	utxo.Change = controlProg.Change

	hash2, err := MustDecodeHash("7bbe6226b226a94e88d15f1d41d28a853059cbdeb9048aad4f3f9243fe953001")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo1 := account.UTXO{}
	utxo1.OutputID = bc.Hash{V0: 1}
	utxo1.SourceID = hash2
	utxo1.AssetID = *consensus.BTMAssetID
	utxo1.Amount = 79000000
	utxo1.SourcePos = 1
	utxo1.ControlProgram = controlProg.ControlProgram
	utxo1.AccountID = controlProg.AccountID
	utxo1.Address = controlProg.Address
	utxo1.ControlProgramIndex = controlProg.KeyIndex
	utxo1.Change = controlProg.Change

	u = append(u, utxo)
	u = append(u, utxo1)

	return u
}

func MakeUTXO(controlProg *account.CtrlProgram) []account.UTXO {
	var u []account.UTXO
	hash, err := MustDecodeHash("d25a673df4dbadcbb9edb0a46a98bcc57bbff6bf2d0b0ac646f44ebb5474b201")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo := account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = hash
	utxo.AssetID = *consensus.BTMAssetID
	utxo.Amount = 100000000
	utxo.SourcePos = 0
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	utxo.Change = controlProg.Change

	hash2, err := MustDecodeHash("7bbe6226b226a94e88d15f1d41d28a853059cbdeb9048aad4f3f9243fe953001")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo1 := account.UTXO{}
	utxo1.OutputID = bc.Hash{V0: 1}
	utxo1.SourceID = hash2
	utxo1.AssetID = *consensus.BTMAssetID
	utxo1.Amount = 79000000
	utxo1.SourcePos = 1
	utxo1.ControlProgram = controlProg.ControlProgram
	utxo1.AccountID = controlProg.AccountID
	utxo1.Address = controlProg.Address
	utxo1.ControlProgramIndex = controlProg.KeyIndex
	utxo1.Change = controlProg.Change

	u = append(u, utxo)
	u = append(u, utxo1)

	return u
}

func MakeUTXO2(controlProg *account.CtrlProgram, sourceID string, amount, sourcePos uint64) *account.UTXO {
	hash, err := MustDecodeHash(sourceID)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	utxo := &account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = hash
	utxo.AssetID = *consensus.BTMAssetID
	utxo.Amount = amount
	utxo.SourcePos = sourcePos
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	utxo.Change = controlProg.Change
	return utxo
}

func MustDecodeHash(s string) (h bc.Hash, err error) {
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return bc.Hash{}, err
	}
	return h, nil
}

// MockTx mock a tx
func MockTx(utxo *account.UTXO, testAccount *account.Account) (*txbuilder.Template, *types.TxData, error) {
	txInput, sigInst, err := account.UtxoToInputs(testAccount.Signer, utxo)
	if err != nil {
		return nil, nil, err
	}

	b := txbuilder.NewBuilder(time.Now())
	if err := b.AddInput(txInput, sigInst); err != nil {
		return nil, nil, err
	}

	cp, err := hexutil.Decode("0x0014053bd6571c2dc924de9f95f2f6a87cec316df1ac")
	out := types.NewOriginalTxOutput(*consensus.BTMAssetID, 99551000, cp, nil)
	if err := b.AddOutput(out); err != nil {
		return nil, nil, err
	}
	return b.Build()
}

func MockTx2(utxo []account.UTXO, testAccount *account.Account) (*txbuilder.Template, *types.TxData, error) {

	b := txbuilder.NewBuilder(time.Now())
	for _, u := range utxo {
		txInput, sigInst, err := account.UtxoToInputs(testAccount.Signer, &u)
		if err != nil {
			return nil, nil, err
		}
		if err := b.AddInput(txInput, sigInst); err != nil {
			return nil, nil, err
		}
	}

	cp, err := hexutil.Decode("0x0014053bd6571c2dc924de9f95f2f6a87cec316df1ac")
	if err != nil {
		return nil, nil, err
	}
	out := types.NewOriginalTxOutput(*consensus.BTMAssetID, 110000000, cp, nil)

	cp2, err := hexutil.Decode("0x00141021717faa4ab1dd52cd65e7325992bbd16f7785")
	if err != nil {
		return nil, nil, err
	}
	out2 := types.NewOriginalTxOutput(*consensus.BTMAssetID, 68102000, cp2, nil)

	if err := b.AddOutput(out); err != nil {
		return nil, nil, err
	}

	if err := b.AddOutput(out2); err != nil {
		return nil, nil, err
	}
	return b.Build()
}

func MakeTx(utxo []account.UTXO, testAccount *account.Account) (*txbuilder.Template, *types.TxData, error) {

	b := txbuilder.NewBuilder(time.Now())
	for _, u := range utxo {
		txInput, sigInst, err := account.UtxoToInputs(testAccount.Signer, &u)
		if err != nil {
			return nil, nil, err
		}
		if err := b.AddInput(txInput, sigInst); err != nil {
			return nil, nil, err
		}
	}

	cp, err := hexutil.Decode("0x0014053bd6571c2dc924de9f95f2f6a87cec316df1ac")
	if err != nil {
		return nil, nil, err
	}
	out := types.NewOriginalTxOutput(*consensus.BTMAssetID, 110000000, cp, nil)
	//070100020160015ee06f1bba024da1cc4b4d9aa34034bbacb5391764a5b67f982b5f64193522d7dfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c2d72f00011600143258626498e281ee85676f4d13e7e2409d73ee12006302401ae70069907e93731b524bd8b9a47e94654434e8e2a3ccbd8770b6ca57a5f2516592033893a8700a966ed4165c80c2686d3496bd4a8df84613881db4d64f490d20212d75b6d6c31c6a8ba59e8e167869edd9eec2b6ccb1398c4032a228aeccbd070160015ecf863ce4447dfeedde72904fae033b7a64c1241bc6eb6cf368647fe953dce331ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c2d72f00011600143cba15f355d63794cf77068db63d3cea880c36dc0063024071f19752c2d7ec1479b198e0268d30500fdaf7650bed9d29d2031c6f124dcabcd6c3b33cc231f242b65d9d7942163bbdf9e05b18a9daa5e8562abea83d36d3012066306e78b47c7d3170ac9be227380b1ff9ed05c7796bb0b107c7603ba2616cf30201003dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff809c9c3901160014e3c308a96552ca01eafb639d80d8d5988b684072000001003dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80bbb02101160014be9c66d5f25d4504435bd45cecda99734df31eb70000
	//070100020160015ee06f1bba024da1cc4b4d9aa34034bbacb5391764a5b67f982b5f64193522d7dfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c2d72f00011600143258626498e281ee85676f4d13e7e2409d73ee12006302401ae70069907e93731b524bd8b9a47e94654434e8e2a3ccbd8770b6ca57a5f2516592033893a8700a966ed4165c80c2686d3496bd4a8df84613881db4d64f490d20212d75b6d6c31c6a8ba59e8e167869edd9eec2b6ccb1398c4032a228aeccbd070160015ecf863ce4447dfeedde72904fae033b7a64c1241bc6eb6cf368647fe953dce331ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80c2d72f00011600143cba15f355d63794cf77068db63d3cea880c36dc0063024071f19752c2d7ec1479b198e0268d30500fdaf7650bed9d29d2031c6f124dcabcd6c3b33cc231f242b65d9d7942163bbdf9e05b18a9daa5e8562abea83d36d3012066306e78b47c7d3170ac9be227380b1ff9ed05c7796bb0b107c7603ba2616cf30201003dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff809c9c3901160014e3c308a96552ca01eafb639d80d8d5988b684072000001003dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80bbb02101160014be9c66d5f25d4504435bd45cecda99734df31eb70000

	cp2, err := hexutil.Decode("0x00141021717faa4ab1dd52cd65e7325992bbd16f7785")
	if err != nil {
		return nil, nil, err
	}
	out2 := types.NewOriginalTxOutput(*consensus.BTMAssetID, 68102000, cp2, nil)

	if err := b.AddOutput(out); err != nil {
		return nil, nil, err
	}

	if err := b.AddOutput(out2); err != nil {
		return nil, nil, err
	}
	return b.Build()
}

func MakeTx2(utxo []*account.UTXO, accounts []*account.Account, outs []*validator.TxOutTpl) (*txbuilder.Template, *types.TxData, error) {
	b := txbuilder.NewBuilder(time.Now())
	for i, u := range utxo {
		txInput, sigInst, err := account.UtxoToInputs(accounts[i].Signer, u)
		if err != nil {
			return nil, nil, err
		}
		if err := b.AddInput(txInput, sigInst); err != nil {
			return nil, nil, err
		}
	}

	for _, o := range outs {
		program, err := AddressToProgram(o.ToAddr)
		if err != nil {
			return nil, nil, err
		}
		//cp, err := hexutil.Decode("0x0014053bd6571c2dc924de9f95f2f6a87cec316df1ac")
		//if err != nil {
		//	return nil, nil, err
		//}
		out := types.NewOriginalTxOutput(*consensus.BTMAssetID, uint64(o.ToAmountInt64), program, nil)
		if err := b.AddOutput(out); err != nil {
			return nil, nil, err
		}
	}

	//cp, err := hexutil.Decode("0x0014053bd6571c2dc924de9f95f2f6a87cec316df1ac")
	//if err != nil {
	//	return nil, nil, err
	//}
	//out := types.NewOriginalTxOutput(*consensus.BTMAssetID, 110000000,cp, nil)

	//cp2, err := hexutil.Decode("0x00141021717faa4ab1dd52cd65e7325992bbd16f7785")
	//if err != nil {
	//	return nil, nil, err
	//}
	//out2 := types.NewOriginalTxOutput(*consensus.BTMAssetID, 68102000,cp2, nil)

	//if err := b.AddOutput(out); err != nil {
	//	return nil, nil, err
	//}
	//
	//if err := b.AddOutput(out2); err != nil {
	//	return nil, nil, err
	//}
	return b.Build()
}

func AddressToProgram(address string) ([]byte, error) {
	addr, err := common.DecodeAddress(address, &consensus.ActiveNetParams)
	if err != nil {
		return nil, err
	}
	redeemContract := addr.ScriptAddress()
	switch addr.(type) {
	case *common.AddressWitnessPubKeyHash:
		program, err := vmutil.P2WPKHProgram(redeemContract)
		return program, err
	case *common.AddressWitnessScriptHash:
		program, err := vmutil.P2WSHProgram(redeemContract)
		return program, err
	default:
		return nil, errors.New("Do not have this type address")
	}
}

// MockSign sign a tx
func MockSign(tpl *txbuilder.Template, hsm *pseudohsm.HSM, password string) (bool, error) {
	err := txbuilder.Sign(nil, tpl, password,
		func(_ context.Context, xpub chainkd.XPub, path [][]byte, data [32]byte, password string) ([]byte, error) {
			return hsm.XSign(xpub, path, data[:], password)
		},
	)
	if err != nil {
		return false, err
	}
	return txbuilder.SignProgress(tpl), nil
}

func MakeSign(tpl *txbuilder.Template, hsm *pseudohsm.HSM, password string) (bool, error) {
	err := txbuilder.Sign(nil, tpl, password,
		func(_ context.Context, xpub chainkd.XPub, path [][]byte, data [32]byte, password string) ([]byte, error) {
			return hsm.XSign(xpub, path, data[:], password)
		},
	)
	if err != nil {
		return false, err
	}
	return txbuilder.SignProgress(tpl), nil
}

// MockBlock mock a block
func MockBlock() *bc.Block {
	return &bc.Block{
		BlockHeader: &bc.BlockHeader{Height: 1},
	}
}
