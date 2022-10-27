package wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"btmSign/bytom/account"
	"btmSign/bytom/asset"
	"btmSign/bytom/blockchain/query"
	"btmSign/bytom/blockchain/signers"
	"btmSign/bytom/common"
	"btmSign/bytom/consensus"
	"btmSign/bytom/consensus/segwit"
	"btmSign/bytom/crypto/sha3pool"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/protocol/bc"
	"btmSign/bytom/protocol/bc/types"
	"btmSign/bytom/protocol/vm/vmutil"
)

// annotateTxs adds asset data to transactions
func annotateTxsAsset(w *Wallet, txs []*query.AnnotatedTx) {
	for i, tx := range txs {
		for j, input := range tx.Inputs {
			alias, definition := w.getAliasDefinition(input.AssetID)
			txs[i].Inputs[j].AssetAlias, txs[i].Inputs[j].AssetDefinition = alias, &definition
		}
		for k, output := range tx.Outputs {
			alias, definition := w.getAliasDefinition(output.AssetID)
			txs[i].Outputs[k].AssetAlias, txs[i].Outputs[k].AssetDefinition = alias, &definition
		}
	}
}

func (w *Wallet) getExternalDefinition(assetID *bc.AssetID) json.RawMessage {
	definitionByte := w.DB.Get(asset.ExtAssetKey(assetID))
	if definitionByte == nil {
		return nil
	}

	definitionMap := make(map[string]interface{})
	if err := json.Unmarshal(definitionByte, &definitionMap); err != nil {
		return nil
	}

	alias := assetID.String()
	externalAsset := &asset.Asset{
		AssetID:           *assetID,
		Alias:             &alias,
		DefinitionMap:     definitionMap,
		RawDefinitionByte: definitionByte,
		Signer:            &signers.Signer{Type: "external"},
	}

	if err := w.AssetReg.SaveAsset(externalAsset, alias); err != nil {
		log.WithFields(log.Fields{"module": logModule, "err": err, "assetID": alias}).Warning("fail on save external asset to internal asset DB")
	}
	return definitionByte
}

func (w *Wallet) getAliasDefinition(assetID bc.AssetID) (string, json.RawMessage) {
	//btm
	if assetID.String() == consensus.BTMAssetID.String() {
		alias := consensus.BTMAlias
		definition := []byte(asset.DefaultNativeAsset.RawDefinitionByte)

		return alias, definition
	}

	//local asset and saved external asset
	if localAsset, err := w.AssetReg.FindByID(nil, &assetID); err == nil {
		alias := *localAsset.Alias
		definition := []byte(localAsset.RawDefinitionByte)
		return alias, definition
	}

	//external asset
	if definition := w.getExternalDefinition(&assetID); definition != nil {
		return assetID.String(), definition
	}

	return "", nil
}

// annotateTxs adds account data to transactions
func annotateTxsAccount(txs []*query.AnnotatedTx, walletDB dbm.DB) {
	for i, tx := range txs {
		for j, input := range tx.Inputs {
			//issue asset tx input SpentOutputID is nil
			if input.SpentOutputID == nil {
				continue
			}
			localAccount, err := getAccountFromACP(input.ControlProgram, walletDB)
			if localAccount == nil || err != nil {
				continue
			}
			txs[i].Inputs[j].AccountAlias = localAccount.Alias
			txs[i].Inputs[j].AccountID = localAccount.ID
		}
		for j, output := range tx.Outputs {
			localAccount, err := getAccountFromACP(output.ControlProgram, walletDB)
			if localAccount == nil || err != nil {
				continue
			}
			txs[i].Outputs[j].AccountAlias = localAccount.Alias
			txs[i].Outputs[j].AccountID = localAccount.ID
		}
	}
}

func getAccountFromACP(program []byte, walletDB dbm.DB) (*account.Account, error) {
	var hash common.Hash
	accountCP := account.CtrlProgram{}
	localAccount := account.Account{}

	sha3pool.Sum256(hash[:], program)

	rawProgram := walletDB.Get(account.ContractKey(hash))
	if rawProgram == nil {
		return nil, fmt.Errorf("failed get account control program:%x ", hash)
	}

	if err := json.Unmarshal(rawProgram, &accountCP); err != nil {
		return nil, err
	}

	accountValue := walletDB.Get(account.Key(accountCP.AccountID))
	if accountValue == nil {
		return nil, fmt.Errorf("failed get account:%s ", accountCP.AccountID)
	}

	if err := json.Unmarshal(accountValue, &localAccount); err != nil {
		return nil, err
	}

	return &localAccount, nil
}

var emptyJSONObject = json.RawMessage(`{}`)

func isValidJSON(b []byte) bool {
	var v interface{}
	err := json.Unmarshal(b, &v)
	return err == nil
}

func (w *Wallet) buildAnnotatedTransaction(orig *types.Tx, b *types.Block, indexInBlock int) *query.AnnotatedTx {
	tx := &query.AnnotatedTx{
		ID:                     orig.ID,
		Timestamp:              b.Timestamp,
		BlockID:                b.Hash(),
		BlockHeight:            b.Height,
		Position:               uint32(indexInBlock),
		BlockTransactionsCount: uint32(len(b.Transactions)),
		Inputs:                 make([]*query.AnnotatedInput, 0, len(orig.Inputs)),
		Outputs:                make([]*query.AnnotatedOutput, 0, len(orig.Outputs)),
		Size:                   orig.SerializedSize,
	}
	for i := range orig.Inputs {
		tx.Inputs = append(tx.Inputs, w.BuildAnnotatedInput(orig, uint32(i)))
	}
	for i := range orig.Outputs {
		tx.Outputs = append(tx.Outputs, w.BuildAnnotatedOutput(orig, i))
	}
	return tx
}

// BuildAnnotatedInput build the annotated input.
func (w *Wallet) BuildAnnotatedInput(tx *types.Tx, i uint32) *query.AnnotatedInput {
	orig := tx.Inputs[i]
	in := &query.AnnotatedInput{
		AssetDefinition: &emptyJSONObject,
	}
	if orig.InputType() != types.CoinbaseInputType {
		in.AssetID = orig.AssetID()
		in.Amount = orig.Amount()
		in.SignData = tx.SigHash(i)
	}

	id := tx.Tx.InputIDs[i]
	in.InputID = id
	e := tx.Entries[id]
	switch e := e.(type) {
	case *bc.Spend:
		in.Type = "spend"
		in.ControlProgram = orig.ControlProgram()
		in.Address = w.getAddressFromControlProgram(in.ControlProgram)
		in.SpentOutputID = e.SpentOutputId
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}
	case *bc.Issuance:
		in.Type = "issue"
		in.IssuanceProgram = orig.ControlProgram()
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}

		if ii, ok := orig.TypedInput.(*types.IssuanceInput); ok && isValidJSON(ii.AssetDefinition) {
			assetDefinition := json.RawMessage(ii.AssetDefinition)
			in.AssetDefinition = &assetDefinition
		}
	case *bc.Coinbase:
		in.Type = "coinbase"
		in.Arbitrary = e.Arbitrary
	case *bc.VetoInput:
		in.Type = "veto"
		in.ControlProgram = orig.ControlProgram()
		in.Address = w.getAddressFromControlProgram(in.ControlProgram)
		in.SpentOutputID = e.SpentOutputId
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}
		if vetoInput, ok := orig.TypedInput.(*types.VetoInput); ok {
			in.Vote = hex.EncodeToString(vetoInput.Vote)
			in.Amount = vetoInput.Amount
		}
	}
	return in
}

func (w *Wallet) getAddressFromControlProgram(prog []byte) string {
	if segwit.IsP2WPKHScript(prog) {
		if pubHash, err := segwit.GetHashFromStandardProg(prog); err == nil {
			return buildP2PKHAddress(pubHash)
		}
	} else if segwit.IsP2WSHScript(prog) {
		if scriptHash, err := segwit.GetHashFromStandardProg(prog); err == nil {
			return buildP2SHAddress(scriptHash)
		}
	}

	return ""
}

func buildP2PKHAddress(pubHash []byte) string {
	address, err := common.NewAddressWitnessPubKeyHash(pubHash, &consensus.ActiveNetParams)
	if err != nil {
		return ""
	}

	return address.EncodeAddress()
}

func buildP2SHAddress(scriptHash []byte) string {
	address, err := common.NewAddressWitnessScriptHash(scriptHash, &consensus.ActiveNetParams)
	if err != nil {
		return ""
	}

	return address.EncodeAddress()
}

// BuildAnnotatedOutput build the annotated output.
func (w *Wallet) BuildAnnotatedOutput(tx *types.Tx, idx int) *query.AnnotatedOutput {
	orig := tx.Outputs[idx]
	outid := tx.OutputID(idx)
	out := &query.AnnotatedOutput{
		OutputID:        *outid,
		Position:        idx,
		AssetID:         *orig.AssetId,
		AssetDefinition: &emptyJSONObject,
		Amount:          orig.Amount,
		ControlProgram:  orig.ControlProgram,
		Address:         w.getAddressFromControlProgram(orig.ControlProgram),
	}

	switch {
	// must deal with retirement first due to cases' priorities in the switch statement
	case vmutil.IsUnspendable(out.ControlProgram):
		// retirement
		out.Type = "retire"
	case orig.OutputType() == types.OriginalOutputType:
		out.Type = "control"
		if e, ok := tx.Entries[*outid]; ok {
			if output, ok := e.(*bc.OriginalOutput); ok {
				out.StateData = stateDataStrings(output.StateData)
			}
		}
	case orig.OutputType() == types.VoteOutputType:
		out.Type = "vote"
		if e, ok := tx.Entries[*outid]; ok {
			if output, ok := e.(*bc.VoteOutput); ok {
				out.Vote = hex.EncodeToString(output.Vote)
				out.StateData = stateDataStrings(output.StateData)
			}
		}
	default:
		log.Warn("unknown outType")
	}

	return out
}

func stateDataStrings(stateData [][]byte) []string {
	ss := make([]string, 0, len(stateData))
	for _, bytes := range stateData {
		ss = append(ss, hex.EncodeToString(bytes))
	}
	return ss
}
