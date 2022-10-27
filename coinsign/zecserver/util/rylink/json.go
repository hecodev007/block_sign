package rylink

import "encoding/json"

func DecodeCreaterawtransactionResult(ds []byte) (*CreaterawtransactionResult, error) {
	ri := &CreaterawtransactionResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeSignResult(ds []byte) (*SignResult, error) {
	ri := &SignResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeSenndTxResult(ds []byte) (*SenndTxResult, error) {
	ri := &SenndTxResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeGetNewAddressResult(ds []byte) (*GetNewAddressResult, error) {
	ri := &GetNewAddressResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeDumpprivkeyResult(ds []byte) (*DumpprivkeyResult, error) {
	ri := &DumpprivkeyResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func DecodeImportprivkeyResult(ds []byte) (*ImportprivkeyResult, error) {
	ri := &ImportprivkeyResult{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}
