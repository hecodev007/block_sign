package v1

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JFJun/rpc-tool/rpc/gxc"
	"github.com/shopspring/decimal"
	"gxclient-go/keypair"
	"gxclient-go/sign"
	"gxclient-go/types"
	"strconv"
	"time"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"
)

type GxcService struct {
	*BaseService
	client *gxc.GXCRpc
}

const (
	GxcDecimal = 5
)

func (bs *BaseService) GXCService() *GxcService {
	gs := new(GxcService)
	gs.BaseService = bs
	gs.client = gxc.NewGXCRpc(conf.Config.GxcCfg.NodeUrl)
	return gs
}

func (gs *GxcService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	return nil, errors.New("unsupport it")
}
func (gs *GxcService) ValidAddress(address string) error {
	_, err := gs.client.GetAccount(address)
	if err != nil {
		return err
	}
	return nil
}
func (gs *GxcService) SignService(req *model.ReqSignParams) (interface{}, error) {
	reqData, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	var signParams model.GxcSignParams
	if err := json.Unmarshal(reqData, &signParams); err != nil {
		return nil, err
	}

	if signParams.PublicKey == "" {
		return nil, errors.New("gxc public key is null")
	}
	if signParams.StxHex == "" {
		return nil, errors.New("gxc stx hex is null")
	}
	wif, err := gs.BaseService.addressOrPublicKeyToPrivate(signParams.PublicKey)
	if err != nil {
		return nil, err
	}
	var chainId string
	if signParams.ChainId == "" {
		chainId = conf.Config.GxcCfg.ChainId
	} else {
		chainId = signParams.ChainId
	}
	//进行签名
	data, _ := hex.DecodeString(signParams.StxHex)
	var stx *types.SignedTransaction
	if err := json.Unmarshal(data, &stx); err != nil {
		return nil, fmt.Errorf("json unmarshal stxHex error,Err=%v", err)
	}
	if stx == nil {
		return nil, errors.New("stx is nil")
	}
	if err := stx.Sign([]string{wif}, chainId); err != nil {
		return nil, fmt.Errorf("sign error,Err=%v", err)
	}
	//将stx转换为hex发送给客户端进行broadcast
	signData, err := json.Marshal(stx)
	if err != nil {
		return nil, fmt.Errorf("json marshal stx error,Err=[%v]", err)
	}
	signHex := hex.EncodeToString(signData)

	return signHex, nil
}
func (gs *GxcService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return gs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, gs.createAddressInfo)
	}
	return gs.BaseService.createAddress(req, gs.createAddressInfo)
	//addresses,err:=util.CreateAddrCsv(conf.Config.FilePath,req.MchId,req.OrderId,req.CoinName,addrInfos)
	//if err != nil {
	//	return nil,err
	//}
	//resp := new(model.RespCreateAddressParams)
	//resp.Address = addresses
	//resp.Num = req.Num
	//resp.OrderId = req.OrderId
	//resp.MchId = req.MchId
	//resp.CoinName = req.CoinName
	//return resp, nil
}
func (gs *GxcService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := gs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, gs.createAddressInfo)
	return err
}

func (gs *GxcService) TransferService(req interface{}) (interface{}, error) {
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var tp model.GxcTransferParams
	if err := json.Unmarshal(reqData, &tp); err != nil {
		return nil, err
	}
	if tp.FromAccount == "" {
		return nil, errors.New("from account  is null")
	}
	if tp.ToAccount == "" {
		return nil, errors.New("to account  is null")
	}
	if tp.Amount == "" {
		return nil, errors.New("amount  is null")
	}
	from, err := gs.client.GetAccount(tp.FromAccount)
	if err != nil {
		return nil, err
	}
	to, err := gs.client.GetAccount(tp.ToAccount)
	if err != nil {
		return nil, err
	}
	amountSymbol, err := gs.client.GetAsset("GXC")
	if err != nil {
		return nil, err
	}

	amt, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("amount(%s) is not str: %v", tp.Amount, err)
	}

	//组合amount
	amountAssets := types.AssetAmount{
		AssetID: amountSymbol.ID,
		Amount:  uint64(amt.Shift(int32(GxcDecimal)).IntPart()),
	}
	//手续费的ID和amount的ID一样都是GXC
	feeAssets := types.AssetAmount{
		AssetID: amountSymbol.ID,
		Amount:  0,
	}

	//组合memo
	var memoOb = &types.Memo{}
	resp := new(model.RespGxcTransferParams)
	if len(tp.Memo) > 0 {

		memoOb.From = from.Options.MemoKey
		memoOb.To = to.Options.MemoKey
		memoOb.Nonce = types.GetNonce()
		private, err := types.NewPrivateKeyFromWif(conf.Config.GxcCfg.MemoKey)
		if err != nil {
			panic(err)
		}
		err = gs.encrypt_memo(memoOb, tp.Memo, private)
		if err != nil {
			return nil, err
		}
		//验证memo是否能解析出来
		if err := gs.validMemo(memoOb, tp.Memo); err != nil {
			return nil, err
		}
		resp.Memo = memoOb.Message.String()
		//memoOb = sc.imitateEncryptMemo(memoOb,memo)
	} else {
		resp.Memo = ""
		memoOb = nil
	}
	//构建Operation
	op := types.NewTransferOperation(types.MustParseObjectID(from.ID.String()),
		types.MustParseObjectID(to.ID.String()), amountAssets, feeAssets, memoOb)

	//计算手续费
	if tp.Fee != "" {
		f := StrToFloat64(tp.Fee)
		ff := f * 100000
		op.Fee.Amount = uint64(ff)
		resp.Fee = tp.Fee
	} else {
		fees, err := gs.client.GetRequiredFee([]types.Operation{op}, feeAssets.AssetID.String())
		if err != nil {
			return nil, err
		}
		op.Fee.Amount = fees[0].Amount
		f := fmt.Sprintf("%.5f", float64(fees[0].Amount)/100000.0)
		resp.Fee = f
	}

	//获取全局属性
	props, err := gs.client.GetDynamicGlobalProperties()
	if err != nil {
		return nil, err
	}
	block, err := gs.client.GetBlock(props.LastIrreversibleBlockNum)
	if err != nil {
		return nil, err
	}
	refBlockPrefix, err := sign.RefBlockPrefix(block.Previous)
	if err != nil {
		return nil, err
	}
	expiration := props.Time.Add(10 * time.Minute)

	//构建签名交易
	stx := types.NewSignedTransaction(&types.Transaction{
		RefBlockNum:    sign.RefBlockNum(props.LastIrreversibleBlockNum - 1&0xffff),
		RefBlockPrefix: refBlockPrefix,
		Expiration:     types.Time{Time: &expiration},
	})
	stx.PushOperation(op)
	//将构建好的签名交易编码成16进制，发送给签名程序
	data, err := json.Marshal(stx)
	if err != nil {
		return nil, err
	}
	stxHex := hex.EncodeToString(data)
	signReqData := &model.GxcSignParams{
		PublicKey: tp.PublicKey,
		StxHex:    stxHex,
		ChainId:   "",
	}
	reqSign := model.ReqSignParams{
		ReqBaseParams: model.ReqBaseParams{},
		Data:          signReqData,
	}
	signResp, err := gs.SignService(&reqSign)
	if err != nil {
		return nil, err
	}
	ss := signResp.(string)
	txid, err := gs.broadcastTransaction(ss)
	if err != nil {
		return nil, err
	}
	if txid == "" {
		resp.Txid = ""
		return nil, errors.New("unknow broadcast tx result")
	}
	resp.Txid = txid
	return resp, nil
}

// string to float64
func StrToFloat64(str string) float64 {
	tmp, _ := strconv.ParseFloat(str, 64)
	return tmp
}

/*
广播交易
*/
func (gs *GxcService) broadcastTransaction(stxHex string) (string, error) {
	data, _ := hex.DecodeString(stxHex)
	var stx *types.SignedTransaction
	if err := json.Unmarshal(data, &stx); err != nil {
		return "", err
	}
	if stx == nil {
		return "", errors.New("broadcast error")
	}
	resp, err := gs.client.BroadcastTransactionSynchronous(stx.Transaction)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (gs *GxcService) encrypt_memo(p *types.Memo, msg string, priv *types.PrivateKey) error {
	sec, err := priv.SharedSecret(&p.To, 16, 16)

	if err != nil {
		return fmt.Errorf("get shared secret error,Err=[%v]", err)
	}

	iv, blk, err := cypherBlock(*p, sec)
	if err != nil {
		return fmt.Errorf("cypherBlock error,Err=[%v]", err)
	}

	buf := []byte(msg)
	digest := sha256.Sum256(buf)
	mode := cipher.NewCBCEncrypter(blk, iv)

	// checksum + msg
	raw := digest[:4]
	raw = append(raw, buf...)

	raw = pad(raw, 16)

	dst := make([]byte, len(raw))
	mode.CryptBlocks(dst, raw)
	p.Message = dst

	return nil
}

func (gs *GxcService) validMemo(memo *types.Memo, memoMsg string) error {
	errs := func(err error) error {

		return fmt.Errorf("Valid memo error,Err=[%v]", err)
	}
	memoKey, err := types.NewPrivateKeyFromWif(conf.Config.GxcCfg.MemoKey)
	if err != nil {

		return errs(err)
	}
	sec, err := memoKey.SharedSecret(&memo.To, 16, 16)
	if err != nil {

		return errs(err)
	}
	iv, blk, err := cypherBlock(*memo, sec)
	if err != nil {

		return errs(err)
	}
	mode := cipher.NewCBCDecrypter(blk, iv)
	dst := make([]byte, len(memo.Message))

	mode.CryptBlocks(dst, memo.Message)

	//verify checksum
	chk1 := dst[:4]
	msg := unpad(dst[4:])
	dig := sha256.Sum256(msg)
	chk2 := dig[:4]
	if bytes.Compare(chk1, chk2) != 0 {
		return errs(errors.New("verify memo checksum error,dont set memo length is [12,28,44]"))
	}
	if string(msg) != memoMsg {
		return errs(fmt.Errorf("memo string is not equal!!!,ParseMemo= %s, memo = %s", string(msg), memoMsg))
	}
	return nil
}

func cypherBlock(p types.Memo, sec []byte) ([]byte, cipher.Block, error) {
	ss := sha512.Sum512(sec)

	var seed []byte
	seed = append(seed, []byte(strconv.FormatUint(uint64(p.Nonce), 10))...)
	seed = append(seed, []byte(hex.EncodeToString(ss[:]))...)

	sd := sha512.Sum512(seed)
	blk, err := aes.NewCipher(sd[0:32])
	if err != nil {
		return nil, nil, err
	}

	return sd[32:48], blk, nil
}

func unpad(buf []byte) []byte {
	b := buf[len(buf)-1:][0]
	cnt := int(b)
	l := len(buf) - cnt
	if l < 0 {
		return make([]byte, 4)
	}
	a := bytes.Repeat([]byte{b}, cnt)
	if bytes.Compare(a, buf[l:]) == 0 {
		return buf[:l]
	}

	return buf
}

func pad(buf []byte, length int) []byte {
	cnt := length - len(buf)%length
	buf = append(buf, bytes.Repeat([]byte{byte(cnt)}, cnt)...)
	return buf
}

/*
创建相应数量的公私钥
*/
func (gs *GxcService) createAddressInfo() (util.AddrInfo, error) {

	var addrInfo util.AddrInfo
	keyPair, err := keypair.GenerateKeyPair("")
	if err != nil {
		return addrInfo, err
	}
	mnemonic := keyPair.BrainKey
	wif := keyPair.PrivateKey.ToWIF()
	pubKey := keyPair.PrivateKey.PublicKey().String()
	addrInfo.Mnemonic = mnemonic
	addrInfo.PrivKey = wif
	addrInfo.Address = pubKey

	return addrInfo, nil
}
