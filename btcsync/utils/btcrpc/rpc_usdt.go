package btcrpc

//根据txid获取详细的的usdt交易信息
func (c *RpcClient) OmniGetrawtransaction(txid string) (*OmniGettransaction, error) {
	var (
		datas []byte
		tx    *OmniGettransaction
		err   error
	)
	datas, err = c.Call("omni_gettransaction", txid)
	if err != nil {
		return nil, err
	}
	tx, err = decodeOmniGetrawtransaction(datas)
	if err != nil {
		return nil, err
	}
	return tx, err
}
