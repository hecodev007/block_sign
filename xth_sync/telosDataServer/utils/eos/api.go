package eos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"telosDataServer/common/log"
)

type API struct {
	HttpClient *http.Client
	BaseURL    string
	Debug      bool
	// Header is one or more headers to be added to all outgoing calls
	Header http.Header
}

func NewAPI(baseURL, apiKey string) *API {
	api := &API{
		HttpClient: http.DefaultClient,
		BaseURL:    baseURL,
		Header:     make(http.Header),
		Debug:      false,
	}

	api.Header.Set("Content-Type", "application/json")
	api.Header.Set("apikey", apiKey)

	return api
}

func (api *API) GetBestHeight() (int64, error) {
	info, err := api.GetInfo()
	if err != nil {
		return -1, err
	}
	return int64(info.LastIrreversibleBlockNum), nil
}

func (api *API) GetInfo() (out *InfoResp, err error) {
	err = api.call("v1/chain/get_info", nil, &out)
	return
}

func (api *API) GetBlockByID(id string) (out *BlockResp, err error) {
	err = api.call("v1/chain/get_block", M{"block_num_or_id": id}, &out)
	return
}

// GetScheduledTransactionsWithBounds returns scheduled transactions within specified bounds
func (api *API) GetScheduledTransactionsWithBounds(lower_bound string, limit uint32) (out *ScheduledTransactionsResp, err error) {
	err = api.call("v1/chain/get_scheduled_transactions", M{"json": true, "lower_bound": lower_bound, "limit": limit}, &out)
	return
}

// GetScheduledTransactions returns the Top 100 scheduled transactions
func (api *API) GetScheduledTransactions() (out *ScheduledTransactionsResp, err error) {
	return api.GetScheduledTransactionsWithBounds("", 100)
}

func (api *API) GetProducers() (out *ProducersResp, err error) {
	/*
		+FC_REFLECT( eosio::chain_apis::read_only::get_producers_params, (json)(lower_bound)(limit) )
		+FC_REFLECT( eosio::chain_apis::read_only::get_producers_result, (rows)(total_producer_vote_weight)(more) ); */
	err = api.call("v1/chain/get_producers", nil, &out)
	return
}

func (api *API) GetBlockByNum(num int64) (out *BlockResp, err error) {
	err = api.call("v1/chain/get_block", M{"block_num_or_id": fmt.Sprintf("%d", num)}, &out)
	//err = api.call("chain", "get_block", M{"block_num_or_id": num}, &out)
	return
}

func (api *API) GetBlockByNumOrID(query int64) (out *BlockResp, err error) {
	err = api.call("v1/chain/get_block", M{"block_num_or_id": query}, &out)
	return
}

func (api *API) GetBlockByNumOrIDRaw(query string) (out interface{}, err error) {
	err = api.call("v1/chain/get_block", M{"block_num_or_id": query}, &out)
	return
}

func (api *API) GetTransaction(id string) (out *TransactionResp, err error) {
	err = api.call("v1/history/get_transaction", M{"id": id}, &out)
	return
}
func (api *API) GetTransactionFromThird(id string) (out *TransactionThirdResp2, err error) {
	err = api.call("v1/history/get_transaction", M{"id": id}, &out)
	return
}

func (api *API) GetTransactionRaw(id string) (out json.RawMessage, err error) {
	err = api.call("v1/history/get_transaction", M{"id": id}, &out)
	return
}

func (api *API) GetActions(params GetActionsRequest) (out *ActionsResp, err error) {
	err = api.call("v1/history/get_actions", params, &out)
	return
}

func (api *API) GetKeyAccounts(publicKey string) (out *KeyAccountsResp, err error) {
	err = api.call("v1/history/get_key_accounts", M{"public_key": publicKey}, &out)
	return
}

func (api *API) GetControlledAccounts(controllingAccount string) (out *ControlledAccountsResp, err error) {
	err = api.call("v1/history/get_controlled_accounts", M{"controlling_account": controllingAccount}, &out)
	return
}

func (api *API) GetTransactions(name AccountName) (out *TransactionsResp, err error) {
	err = api.call("v1/account_history/get_transactions", M{"account_name": name}, &out)
	return
}

func (api *API) GetTableByScope(params GetTableByScopeRequest) (out *GetTableByScopeResp, err error) {
	err = api.call("v1/chain/get_table_by_scope", params, &out)
	return
}

func (api *API) GetTableRows(params GetTableRowsRequest) (out *GetTableRowsResp, err error) {
	err = api.call("v1/chain/get_table_rows", params, &out)
	return
}

func (api *API) GetRawABI(params GetRawABIRequest) (out *GetRawABIResp, err error) {
	err = api.call("v1/chain/get_raw_abi", params, &out)
	return
}

func (api *API) GetCurrencyBalance(account AccountName, symbol string, code AccountName) (out []Asset, err error) {
	params := M{"account": account, "code": code}
	if symbol != "" {
		params["symbol"] = symbol
	}
	err = api.call("v1/chain/get_currency_balance", params, &out)
	return
}

func (api *API) GetCurrencyStats(code AccountName, symbol string) (out *GetCurrencyStatsResp, err error) {
	params := M{"code": code, "symbol": symbol}

	outWrapper := make(map[string]*GetCurrencyStatsResp)
	err = api.call("v1/chain/get_currency_stats", params, &outWrapper)
	out = outWrapper[symbol]

	return
}

// See more here: libraries/chain/contracts/abi_serializer.cpp:58...
func (api *API) call(baseAPI string, body interface{}, out interface{}) (err error) {
	defer func() {
		if err != nil {
			log.Error(err.Error())
		}
	}()
	jsonBody, err := enc(body)
	if err != nil {
		return err
	}

	targetURL := fmt.Sprintf("%s/%s", api.BaseURL, baseAPI)
	req, err := http.NewRequest("POST", targetURL, jsonBody)
	if err != nil {
		return fmt.Errorf("NewRequest: %s", err)
	}

	for k, v := range api.Header {
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header[k] = append(req.Header[k], v...)
	}

	if api.Debug {
		// Useful when debugging API calls
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(string(requestDump))
	}

	resp, err := api.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %s", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("copy: %s", err)
	}
	if resp.StatusCode == 404 {
		var apiErr APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return ErrNotFound
		}
		return apiErr
	}

	if resp.StatusCode > 299 {
		var apiErr APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
		}
		fmt.Println(string(cnt.Bytes()))
		// Handle cases where some API calls (/v1/chain/get_account for example) returns a 500
		// error when retrieving data that does not exist.
		if apiErr.IsUnknownKeyError() {
			return ErrNotFound
		}

		return apiErr
	}

	if api.Debug {
		fmt.Println("RESPONSE:")
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Printf("%q\n", responseDump)
		fmt.Println("-------------------------------")
	}

	if err := json.Unmarshal(cnt.Bytes(), &out); err != nil {

		return fmt.Errorf("unmarshal: %s", err)
	}

	return nil
}

var ErrNotFound = errors.New("resource not found")

type M map[string]interface{}

func enc(v interface{}) (io.Reader, error) {
	if v == nil {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func (api *API) EnableKeepAlives() bool {
	if tr, ok := api.HttpClient.Transport.(*http.Transport); ok {
		tr.DisableKeepAlives = false
		return true
	}
	return false
}

//func (api *API) GetDBSize() (out *DBSizeResp, err error) {
//	err = api.call("db_size/get", nil, &out)
//	return
//}
//
//func (api *API) SetCustomGetRequiredKeys(f func(tx *Transaction) ([]ecc.PublicKey, error)) {
//	api.customGetRequiredKeys = f
//}
//
//func (api *API) GetAccount(name AccountName) (out *AccountResp, err error) {
//	err = api.call("v1/chain/get_account", M{"account_name": name}, &out)
//	return
//}
//
//func (api *API) GetRawCodeAndABI(account AccountName) (out *GetRawCodeAndABIResp, err error) {
//	err = api.call("v1/chain/get_raw_code_and_abi", M{"account_name": account}, &out)
//	return
//}
//
//func (api *API) GetCode(account AccountName) (out *GetCodeResp, err error) {
//	err = api.call("v1/chain/get_code", M{"account_name": account, "code_as_wasm": true}, &out)
//	return
//}
//
//func (api *API) GetCodeHash(account AccountName) (out Checksum256, err error) {
//	resp := GetCodeHashResp{}
//	if err = api.call("v1/chain/get_code_hash", M{"account_name": account}, &resp); err != nil {
//		return
//	}
//
//	buffer, err := hex.DecodeString(resp.CodeHash)
//	return Checksum256(buffer), err
//}
//
//func (api *API) GetABI(account AccountName) (out *GetABIResp, err error) {
//	err = api.call("v1/chain/get_abi", M{"account_name": account}, &out)
//	return
//}
//
//func (api *API) ABIJSONToBin(code AccountName, action Name, payload M) (out HexBytes, err error) {
//	resp := ABIJSONToBinResp{}
//	err = api.call("v1/chain/abi_json_to_bin", M{"code": code, "action": action, "args": payload}, &resp)
//	if err != nil {
//		return
//	}
//
//	buffer, err := hex.DecodeString(resp.Binargs)
//	return HexBytes(buffer), err
//}
//
//func (api *API) ABIBinToJSON(code AccountName, action Name, payload HexBytes) (out M, err error) {
//	resp := ABIBinToJSONResp{}
//	err = api.call("v1/chain/abi_bin_to_json", M{"code": code, "action": action, "binargs": payload}, &resp)
//	if err != nil {
//		return
//	}
//
//	return resp.Args, nil
//}
//// PushTransaction submits a properly filled (tapos), packed and
//// signed transaction to the blockchain.
//func (api *API) PushTransaction(tx *PackedTransaction) (out *PushTransactionFullResp, err error) {
//	err = api.call("v1/chain/push_transaction", tx, &out)
//	return
//}
//
//func (api *API) PushTransactionRaw(tx *PackedTransaction) (out json.RawMessage, err error) {
//	err = api.call("v1/chain/push_transaction", tx, &out)
//	return
//}
//
//func (api *API) GetNetConnections() (out []*NetConnectionsResp, err error) {
//	err = api.call("v1/net/connections", nil, &out)
//	return
//}
//
//func (api *API) NetConnect(host string) (out NetConnectResp, err error) {
//	err = api.call("v1/net/connect", host, &out)
//	return
//}
//
//func (api *API) NetDisconnect(host string) (out NetDisconnectResp, err error) {
//	err = api.call("v1/net/disconnect", host, &out)
//	return
//}
//
//func (api *API) GetNetStatus(host string) (out *NetStatusResp, err error) {
//	err = api.call("v1/net/status", M{"host": host}, &out)
//	return
//}
//
//// ProducerPause will pause block production on a nodeos with
//// `producer_api` plugin loaded.
//func (api *API) ProducerPause() error {
//	return api.call("v1/producer/pause", nil, nil)
//}
//// CreateSnapshot will write a snapshot file on a nodeos with
//// `producer_api` plugin loaded.
//func (api *API) CreateSnapshot() (out *CreateSnapshotResp, err error) {
//	err = api.call("v1/producer/create_snapshot", nil, &out)
//	return
//}
//// GetIntegrityHash will produce a hash corresponding to current
//// state. Requires `producer_api` and useful when loading
//// from a snapshot
//func (api *API) GetIntegrityHash() (out *GetIntegrityHashResp, err error) {
//	err = api.call("v1/producer/get_integrity_hash", nil, &out)
//	return
//}
//// ProducerResume will resume block production on a nodeos with
//// `producer_api` plugin loaded. Obviously, this needs to be a
//// producing node on the producers schedule for it to do anything.
//func (api *API) ProducerResume() error {
//	return api.call("v1/producer/resume", nil, nil)
//}
//// IsProducerPaused queries the blockchain for the pause statement of
//// block production.
//func (api *API) IsProducerPaused() (out bool, err error) {
//	err = api.call("v1/producer/paused", nil, &out)
//	return
//}
// FixKeepAlives tests the remote server for keepalive support (the
// kava-data `nodeos` software doesn't in the version from March 22nd
// 2018).  Some endpoints front their node with a keep-alive
// supporting web server.  Adjust the `KeepAlive` support of the
// client accordingly.
//func (api *API) FixKeepAlives() bool {
//	// Yeah, to provoke a keep alive, you need to query twice.
//	for i := 0; i < 5; i++ {
//		_, err := api.GetInfo()
//		if api.Debug {
//			log.Println("err", err)
//		}
//		if err == io.EOF {
//			if tr, ok := api.HttpClient.Transport.(*http.Transport); ok {
//				tr.DisableKeepAlives = true
//				return true
//			}
//		}
//		_, err = api.GetNetConnections()
//		if api.Debug {
//			log.Println("err", err)
//		}
//		if err == io.EOF {
//			if tr, ok := api.HttpClient.Transport.(*http.Transport); ok {
//				tr.DisableKeepAlives = true
//				return true
//			}
//		}
//	}
//	return false
//}
