package net

const (
	CreateKey              = "/create-key"
	CreateAccount          = "/create-account"
	CreateAccountReceiver  = "/create-account-receiver"
	ListUnspentOutputs     = "/list-unspent-outputs"
	BuildTransaction       = "/build-transaction"
	BuildChainTransaction  = "/build-chain-transactions"
	SignTransaction        = "/sign-transaction"
	SubmitTransaction      = "/submit-transaction"
	EstimateTransactionGas = "/estimate-transaction-gas"
	ListBalances           = "/list-balances"
)

type SubmitTransactionRequest struct {
	RawTransaction string `json:"raw_transaction"`
}

type SubmitTransactionResult struct {
	Status string `json:"status"`
	Data   struct {
		TxID string `json:"tx_id"`
	} `json:"data"`
}

type ListBalancesRequest struct {
	AccountID string `json:"account_id"`
}

type ListBalancesResult struct {
	Status string `json:"status"`
	Data   []struct {
		AccountID       string `json:"account_id"`
		AccountAlias    string `json:"account_alias"`
		AssetAlias      string `json:"asset_alias"`
		AssetID         string `json:"asset_id"`
		Amount          int64  `json:"amount"`
		AssetDefinition struct {
			Decimals    int    `json:"decimals"`
			Description string `json:"description"`
			Name        string `json:"name"`
			Symbol      string `json:"symbol"`
		} `json:"asset_definition"`
	} `json:"data"`
}

type EstimateTransactionGasRequest struct {
	TransactionTemplate Data `json:"transaction_template"`
}

type EstimateTransactionGasResult struct {
	Status string `json:"status"`
	Data   struct {
		StorageNeu int64 `json:"storage_neu"`
		TotalNeu   int64 `json:"total_neu"`
		VMNeu      int64 `json:"vm_neu"`
	} `json:"data"`
}

type ListUnspentOutputsRequest struct {
	AccountId string `json:"account_id"`
}

type ListUnspentOutputsResult struct {
	Status string `json:"status"`
	Data   []struct {
		AccountAlias        string `json:"account_alias"`
		ID                  string `json:"id"`
		AssetID             string `json:"asset_id"`
		AssetAlias          string `json:"asset_alias"`
		Amount              int64  `json:"amount"`
		AccountID           string `json:"account_id"`
		Address             string `json:"address"`
		ControlProgramIndex int    `json:"control_program_index"`
		Program             string `json:"program"`
		SourceID            string `json:"source_id"`
		SourcePos           int    `json:"source_pos"`
		ValidHeight         int    `json:"valid_height"`
		Change              bool   `json:"change"`
		DeriveRule          int    `json:"derive_rule"`
	} `json:"data"`
}

type SignTransactionRequest struct {
	Password    string `json:"password"`
	Transaction Data   `json:"transaction"`
}

type SignTransactionResult struct {
	Status string `json:"status"`
	Data   struct {
		Transaction struct {
			RawTransaction      string `json:"raw_transaction"`
			SigningInstructions []struct {
				Position          int `json:"position"`
				WitnessComponents []struct {
					Type   string `json:"type"`
					Quorum int    `json:"quorum,omitempty"`
					Keys   []struct {
						Xpub           string   `json:"xpub"`
						DerivationPath []string `json:"derivation_path"`
					} `json:"keys,omitempty"`
					Signatures []string `json:"signatures,omitempty"`
					Value      string   `json:"value,omitempty"`
				} `json:"witness_components"`
			} `json:"signing_instructions"`
			Fee                    int  `json:"fee"`
			AllowAdditionalActions bool `json:"allow_additional_actions"`
		} `json:"transaction"`
		SignComplete bool `json:"sign_complete"`
	} `json:"data"`
}

type BuildTransactionRequest struct {
	BaseTransaction interface{} `json:"base_transaction"`
	Actions         []Action    `json:"actions"`
	TTL             int         `json:"ttl"`
	TimeRange       int         `json:"time_range"`
}

type Action struct {
	AccountID string `json:"account_id,omitempty"`
	Amount    int64  `json:"amount"`
	AssetID   string `json:"asset_id"`
	Type      string `json:"type"`
	Address   string `json:"address,omitempty"`
}

type BuildTransactionResult struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	RawTransaction      string `json:"raw_transaction"`
	SigningInstructions []struct {
		Position          int `json:"position"`
		WitnessComponents []struct {
			Type   string `json:"type"`
			Quorum int    `json:"quorum,omitempty"`
			Keys   []struct {
				Xpub           string   `json:"xpub"`
				DerivationPath []string `json:"derivation_path"`
			} `json:"keys,omitempty"`
			Signatures interface{} `json:"signatures,omitempty"`
			Value      string      `json:"value,omitempty"`
		} `json:"witness_components"`
	} `json:"signing_instructions"`
	Fee                    int  `json:"fee"`
	AllowAdditionalActions bool `json:"allow_additional_actions"`
}

//type BuildTransactionResult struct {
//	Status string `json:"status"`
//	Data   []struct {
//		AccountAlias        string `json:"account_alias"`
//		ID                  string `json:"id"`
//		AssetID             string `json:"asset_id"`
//		AssetAlias          string `json:"asset_alias"`
//		Amount              int    `json:"amount"`
//		AccountID           string `json:"account_id"`
//		Address             string `json:"address"`
//		ControlProgramIndex int    `json:"control_program_index"`
//		Program             string `json:"program"`
//		SourceID            string `json:"source_id"`
//		SourcePos           int    `json:"source_pos"`
//		ValidHeight         int    `json:"valid_height"`
//		Change              bool   `json:"change"`
//		DeriveRule          int    `json:"derive_rule"`
//	} `json:"data"`
//}

type CreateAccountReceiverRequest struct {
	AccountAlias string `json:"account_alias"`
	AccountID    string `json:"account_id"`
}

type CreateAccountReceiverResult struct {
	Status string `json:"status"`
	Data   struct {
		ControlProgram string `json:"control_program"`
		Address        string `json:"address"`
	} `json:"data"`
}

type CreateAccountRequest struct {
	RootXpubs []string `json:"root_xpubs"`
	Quorum    int      `json:"quorum"`
	Alias     string   `json:"alias"`
}

type CreateAccountResult struct {
	Status string `json:"status"`
	Data   struct {
		ID         string   `json:"id"`
		Alias      string   `json:"alias"`
		Xpubs      []string `json:"xpubs"`
		Quorum     int      `json:"quorum"`
		KeyIndex   int      `json:"key_index"`
		DeriveRule int      `json:"derive_rule"`
	} `json:"data"`
}

type CreateKeyRequest struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
	Language string `json:"language"`
}

type CreateKeyResult struct {
	Status string `json:"status"`
	Data   struct {
		Alias    string `json:"alias"`
		Xpub     string `json:"xpub"`
		File     string `json:"file"`
		Mnemonic string `json:"mnemonic"`
	} `json:"data"`
}
