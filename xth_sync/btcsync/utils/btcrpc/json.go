package btcrpc

import "encoding/json"

func decodeBtcGetrawtransaction(ds []byte) (*BtcTx, error) {
	ri := &BtcTx{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

//解析块信息
func decodeBtcBlock(ds []byte) (*BtcBlock, error) {
	ri := &BtcBlock{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBtcGetBlockHash(ds []byte) (*BtcGetBlockHash, error) {
	ri := &BtcGetBlockHash{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBtcChaininfo(ds []byte) (*BtcChaininfo, error) {
	ri := &BtcChaininfo{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBtcBlockCountInfo(ds []byte) (*BtcBlockCountInfo, error) {
	ri := &BtcBlockCountInfo{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeOmniGetrawtransaction(ds []byte) (*OmniGettransaction, error) {
	ri := &OmniGettransaction{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBtcBlockHeader(ds []byte) (*BtcBlockHeader, error) {
	ri := &BtcBlockHeader{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}
