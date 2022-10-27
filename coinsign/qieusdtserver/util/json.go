package util

import (
	"errors"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/json-iterator/go"
)

var (
	json = jsoniter.Config{
		EscapeHTML:                    false,
		MarshalFloatWith6Digits:       false, // will lose precession
		ObjectFieldMustBeSimpleString: true,  // do not unescape object field
	}.Froze()
)

type RespJson struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
type RespJsonData struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Hash    interface{} `json:"hash,omitempty"`
}

func EncodeHttpResopne(code int, msg []byte, data map[string]interface{}) []byte {
	ds, _ := json.Marshal(&RespJson{
		Code:    code,
		Message: string(msg),
		Data:    data,
	})
	return ds

}
func EncodeHttpResopneData(code int, msg []byte, data interface{}) []byte {
	ds, _ := json.Marshal(&RespJsonData{
		Code:    code,
		Message: string(msg),
		Data:    data,
	})
	return ds
}

func EncodeHttpResopneHash(code int, msg []byte, data interface{}, hash string) []byte {
	ds, _ := json.Marshal(&RespJsonData{
		Code:    code,
		Message: string(msg),
		Data:    data,
		Hash:    hash,
	})
	return ds
}

func EncodeSignInputs(sis []*models.SignInput) ([]byte, error) {
	return json.Marshal(sis)
}

func EncodePushInputs(pis []*models.PushInput) ([]byte, error) {
	return json.Marshal(pis)
}

func DecodeStringArray(ds []byte) ([]string, error) {
	arr := make([]string, 0)
	err := json.Unmarshal(ds, &arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

//解码参数
func DecodeAddrTxin(ds []byte) (*models.AddrTxin, error) {
	ti := &models.AddrTxin{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//解码参数
func DecodeAddrTxinFee(ds []byte) (*models.AddrTxinUseFee, error) {
	ti := &models.AddrTxinUseFee{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//解码参数
func DecodeTxInput(ds []byte) (*models.TxInput, error) {
	ti := &models.TxInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//解码参数
func DecodeTxInputNew(ds []byte) (*models.TxInputNew, error) {
	ti := &models.TxInputNew{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func DecodeSignInput(ds []byte) (*models.SignInput, error) {
	ti := &models.SignInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func DecodeHash(ds []byte) (*models.HashInput, error) {
	ti := &models.HashInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func DecodeSignInputNew(ds []byte) ([]*models.SignInput, error) {
	tis := make([]*models.SignInput, 0, 300)
	err := json.Unmarshal(ds, &tis)
	if err != nil {
		return nil, err
	}
	return tis, nil
}

func DecodeSignInputNewOne(ds []byte) (*models.SignInput, error) {
	pi := &models.SignInput{}
	err := json.Unmarshal(ds, &pi)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func DecodePushInput(ds []byte) (*models.PushInput, error) {
	pi := &models.PushInput{}
	err := json.Unmarshal(ds, &pi)
	if err != nil {
		return nil, err
	}
	return pi, nil
}
func DecodePushInputNew(ds []byte) ([]*models.PushInput, error) {
	tis := make([]*models.PushInput, 0, 300)
	err := json.Unmarshal(ds, &tis)
	if err != nil {
		return nil, err
	}
	return tis, nil
}

func DecodeGasInput(ds []byte) (*models.GasInput, error) {
	ti := &models.GasInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//解码参数
func DecodeCreateAddress(ds []byte) (*models.AddressInput, error) {
	ti := &models.AddressInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

//解码参数
func DecodeBatchCreateAddress(ds []byte) (*models.BatchAddressInput, error) {
	ti := &models.BatchAddressInput{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	if ti.Num <= 0 {
		return nil, errors.New("num can not <= 0")
	}

	return ti, nil
}

func DecodeGenNum(ds []byte) (int64, error) {
	var err error
	m := map[string]int64{}
	if err = json.Unmarshal(ds, &m); err != nil {
		return 0, err
	}
	num, ok := m["num"]
	if !ok || num <= 0 {
		return 0, errors.New("num can not <= 0")
	}
	return num, nil
}

func DecodeTxInputs(ds []byte) ([]*models.TxInput, error) {
	tis := make([]*models.TxInput, 0, 300)
	err := json.Unmarshal(ds, &tis)
	if err != nil {
		return nil, err
	}
	return tis, nil
}

func DecodeSignInputs(ds []byte) ([]*models.SignInput, error) {
	sis := make([]*models.SignInput, 0, 300)
	err := json.Unmarshal(ds, &sis)
	if err != nil {
		return nil, err
	}
	return sis, nil
}

func DecodePushInputs(ds []byte) ([]*models.PushInput, error) {
	pis := make([]*models.PushInput, 0, 300)
	err := json.Unmarshal(ds, &pis)
	if err != nil {
		return nil, err
	}
	return pis, nil
}

func DecodeImportKey(ds []byte) (*models.ImportKey, error) {
	ti := &models.ImportKey{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func DecodeImportKey2(ds []byte) (*models.ImportKey2, error) {
	ti := &models.ImportKey2{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func DecodeRemoveKeyInput(ds []byte) (*models.RemoveKeyInput, error) {
	ri := &models.RemoveKeyInput{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeGasHttpResult(ds []byte) (*models.GasHttpResult, error) {
	ri := &models.GasHttpResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeBalanceOutput(ds []byte) (*models.BalanceOutput, error) {
	ri := &models.BalanceOutput{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniSimpleSendResult(ds []byte) (*models.OmniSimpleSendResult, error) {
	ri := &models.OmniSimpleSendResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniCreaterawtransactionResult(ds []byte) (*models.OmniCreaterawtransactionResult, error) {
	ri := &models.OmniCreaterawtransactionResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniOpreturnResult(ds []byte) (*models.OmniOpreturnResult, error) {
	ri := &models.OmniOpreturnResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniReferenceResult(ds []byte) (*models.OmniReferenceResult, error) {
	ri := &models.OmniReferenceResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniChangeResult(ds []byte) (*models.OmniChangeResult, error) {
	ri := &models.OmniChangeResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniSignResult(ds []byte) (*models.OmniSignResult, error) {
	ri := &models.OmniSignResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniSenndTxResult(ds []byte) (*models.OmniSenndTxResult, error) {
	ri := &models.OmniSenndTxResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniGetNewAddressResult(ds []byte) (*models.OmniGetNewAddressResult, error) {
	ri := &models.OmniGetNewAddressResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniDumpprivkeyResult(ds []byte) (*models.OmniDumpprivkeyResult, error) {
	ri := &models.OmniDumpprivkeyResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniValidateaddressResult(ds []byte) (*models.OmniValidateaddressResult, error) {
	ri := &models.OmniValidateaddressResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeOmniImportprivkeyResult(ds []byte) (*models.OmniImportprivkeyResult, error) {
	ri := &models.OmniImportprivkeyResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeImportaddrResult(ds []byte) (*models.ImportaddrResult, error) {
	ti := &models.ImportaddrResult{}
	err := json.Unmarshal(ds, &ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}
