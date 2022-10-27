package bch

import (
	"encoding/json"
	"log"
)

func decodeBchGetrawtransaction(ds []byte) (*BchTx, error) {
	ri := &BchTx{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

//解析块信息
func decodeBchBlock(ds []byte) (*BchBlock, error) {
	ri := &BchBlock{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBchGetBlockHash(ds []byte) (*BchGetBlockHash, error) {
	ri := &BchGetBlockHash{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBchChaininfo(ds []byte) (*BchChaininfo, error) {
	ri := &BchChaininfo{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func decodeBchBlockCountInfo(ds []byte) (*BchBlockCountInfo, error) {
	ri := &BchBlockCountInfo{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		log.Println(err.Error())
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

func decodeBchBlockHeader(ds []byte) (*BchBlockHeader, error) {
	ri := &BchBlockHeader{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

//解析块信息
func decodeBchBlockOnHasTxId(ds []byte) (*BchBlockOnHasTxId, error) {
	ri := &BchBlockOnHasTxId{}
	err := json.Unmarshal(ds, &ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}
