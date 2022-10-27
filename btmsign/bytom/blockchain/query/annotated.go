package query

import (
	"encoding/json"

	"btmSign/bytom/crypto/ed25519/chainkd"
	chainjson "btmSign/bytom/encoding/json"
	"btmSign/bytom/protocol/bc"
)

//AnnotatedTx means an annotated transaction.
type AnnotatedTx struct {
	ID                     bc.Hash            `json:"tx_id"`
	Timestamp              uint64             `json:"block_time"`
	BlockID                bc.Hash            `json:"block_hash"`
	BlockHeight            uint64             `json:"block_height"`
	Position               uint32             `json:"block_index"`
	BlockTransactionsCount uint32             `json:"block_transactions_count,omitempty"`
	Inputs                 []*AnnotatedInput  `json:"inputs"`
	Outputs                []*AnnotatedOutput `json:"outputs"`
	Size                   uint64             `json:"size"`
}

//AnnotatedInput means an annotated transaction input.
type AnnotatedInput struct {
	Type             string               `json:"type"`
	AssetID          bc.AssetID           `json:"asset_id"`
	AssetAlias       string               `json:"asset_alias,omitempty"`
	AssetDefinition  *json.RawMessage     `json:"asset_definition,omitempty"`
	Amount           uint64               `json:"amount"`
	IssuanceProgram  chainjson.HexBytes   `json:"issuance_program,omitempty"`
	ControlProgram   chainjson.HexBytes   `json:"control_program,omitempty"`
	Address          string               `json:"address,omitempty"`
	SpentOutputID    *bc.Hash             `json:"spent_output_id,omitempty"`
	AccountID        string               `json:"account_id,omitempty"`
	AccountAlias     string               `json:"account_alias,omitempty"`
	Arbitrary        chainjson.HexBytes   `json:"arbitrary,omitempty"`
	InputID          bc.Hash              `json:"input_id"`
	WitnessArguments []chainjson.HexBytes `json:"witness_arguments"`
	SignData         bc.Hash              `json:"sign_data,omitempty"`

	// Vote assign value only input is vote type
	Vote      string   `json:"vote,omitempty"`
	StateData []string `json:"state_data,omitempty"`
}

//AnnotatedOutput means an annotated transaction output.
type AnnotatedOutput struct {
	Type            string             `json:"type"`
	OutputID        bc.Hash            `json:"id"`
	TransactionID   *bc.Hash           `json:"transaction_id,omitempty"`
	Position        int                `json:"position"`
	AssetID         bc.AssetID         `json:"asset_id"`
	AssetAlias      string             `json:"asset_alias,omitempty"`
	AssetDefinition *json.RawMessage   `json:"asset_definition,omitempty"`
	Amount          uint64             `json:"amount"`
	AccountID       string             `json:"account_id,omitempty"`
	AccountAlias    string             `json:"account_alias,omitempty"`
	ControlProgram  chainjson.HexBytes `json:"control_program"`
	Address         string             `json:"address,omitempty"`
	// assign value only output is vote type
	Vote string `json:"vote,omitempty"`

	// assign value when output is not retirement type
	StateData []string `json:"state_data,omitempty"`
}

//AnnotatedAccount means an annotated account.
type AnnotatedAccount struct {
	ID         string         `json:"id"`
	Alias      string         `json:"alias,omitempty"`
	XPubs      []chainkd.XPub `json:"xpubs"`
	Quorum     int            `json:"quorum"`
	KeyIndex   uint64         `json:"key_index"`
	DeriveRule uint8          `json:"derive_rule"`
}

//AnnotatedAsset means an annotated asset.
type AnnotatedAsset struct {
	AnnotatedSigner
	ID                bc.AssetID         `json:"id"`
	Alias             string             `json:"alias"`
	VMVersion         uint64             `json:"vm_version"`
	IssuanceProgram   chainjson.HexBytes `json:"issue_program"`
	RawDefinitionByte chainjson.HexBytes `json:"raw_definition_byte"`
	Definition        *json.RawMessage   `json:"definition"`
	LimitHeight       uint64             `json:"limit_height"`
}

//AnnotatedSigner means an annotated signer for asset.
type AnnotatedSigner struct {
	Type       string         `json:"type"`
	XPubs      []chainkd.XPub `json:"xpubs"`
	Quorum     int            `json:"quorum"`
	KeyIndex   uint64         `json:"key_index"`
	DeriveRule uint8          `json:"derive_rule"`
}

//AnnotatedUTXO means an annotated utxo.
type AnnotatedUTXO struct {
	Alias               string `json:"account_alias"`
	OutputID            string `json:"id"`
	AssetID             string `json:"asset_id"`
	AssetAlias          string `json:"asset_alias"`
	Amount              uint64 `json:"amount"`
	AccountID           string `json:"account_id"`
	Address             string `json:"address"`
	ControlProgramIndex uint64 `json:"control_program_index"`
	Program             string `json:"program"`
	SourceID            string `json:"source_id"`
	SourcePos           uint64 `json:"source_pos"`
	ValidHeight         uint64 `json:"valid_height"`
	Change              bool   `json:"change"`
	DeriveRule          uint8  `json:"derive_rule"`
}
