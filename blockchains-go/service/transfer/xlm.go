package transfer

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

type XlmTransferService struct {
	CoinName string
}

func NewXlmTransferService() service.TransferService {
	return &XlmTransferService{CoinName: "xlm"}
}

func (srv *XlmTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		mch        *entity.FcMch
		coinSet    *entity.FcCoinSet
		orderReq   *transfer.XlmOrderHotRequest
		amount     decimal.Decimal //发送金额
		createData []byte          //构造交易信息
	)
	mch, err = dao.FcMchFindById(ta.AppId)
	if err != nil {
		return "", err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coinSet = global.CoinDecimal[coinType]
	if coinSet == nil {
		return "", fmt.Errorf("缺少币种信息")
	}
	orderReq, err = srv.buildOrderHot(ta, int32(coinSet.Decimal))
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	amount, _ = decimal.NewFromString(orderReq.Amount)

	createData, _ = json.Marshal(orderReq)
	orderHot := &entity.FcOrderHot{
		ApplyId:      ta.Id,
		ApplyCoinId:  coinSet.Id,
		OuterOrderNo: ta.OutOrderid,
		OrderNo:      ta.OrderId,
		MchName:      mch.Platform,
		CoinName:     ta.CoinName,
		FromAddress:  orderReq.FromAddress,
		ToAddress:    orderReq.ToAddress,
		Amount:       amount.IntPart(), //转换整型
		Quantity:     amount.String(),
		Decimal:      int64(coinSet.Decimal),
		CreateData:   string(createData),
		Status:       int(status.UnknowErrorStatus),
		CreateAt:     time.Now().Unix(),
		UpdateAt:     time.Now().Unix(),
		Memo:         orderReq.Memo,
		Token:        orderReq.Token,
	}
	txid, err = srv.walletServerCreateHot(orderReq)
	if err != nil {
		orderHot.Status = int(status.BroadcastErrorStatus)
		orderHot.ErrorMsg = err.Error()
		dao.FcOrderHotInsert(orderHot)
		log.Errorf("下单表订单id：%d,获取发送交易异常:%s", ta.Id, err.Error())
		// 写入热钱包表，创建失败
		return "", err
	}
	orderHot.TxId = txid
	orderHot.Status = int(status.BroadcastStatus)
	// 保存热表
	err = dao.FcOrderHotInsert(orderHot)
	if err != nil {
		err = fmt.Errorf("保存订单[%s]数据异常:[%s]", orderHot.OuterOrderNo, err.Error())
		// 保存手续费异常,但是不能中断返回txid，需要回调数据，后续补数据
		log.Error(err.Error())
		// 发送给钉钉
		dingding.ErrTransferDingBot.NotifyStr(err.Error())
	}
	return txid, nil
}
func (srv *XlmTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	orderReq, err := srv.buildOrder(ta)
	if err != nil {
		return err
	}
	err = srv.walletServerCreateCold(orderReq)
	if err != nil {
		//改变表状态
		//7 构建成功
		//8 构建失败，等待重试
		//9 构建失败，不再重试
		return err
	}
	return nil
}

/*
reference； github.com/stellar/go/xdr/muxed_account.go -->method:SetAddress(address string)error
*/
func (srv *XlmTransferService) VaildAddr(address string) error {
	switch len(address) {
	case 56:
		raw, err := decode(VersionByteAccountID, address)
		if err != nil {
			return err
		}
		if len(raw) != 32 {
			return errors.New("invalid address")
		}
		return nil
	case 69:
		raw, err := decode(VersionByteMuxedAccount, address)
		if err != nil {
			return err
		}
		if len(raw) != 40 {
			return errors.New("invalid muxed address")
		}
		return nil
	default:
		return errors.New("invalid address")
	}
}

//创建交易接口参数
func (srv *XlmTransferService) walletServerCreateCold(orderReq *transfer.XlmOrderReq) error {
	data, err := util.PostJsonByAuth(conf.Cfg.Walletserver.Url+"/xlm/create", conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, orderReq)
	if err != nil {
		return err
	}
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	return nil
}

func (srv *XlmTransferService) buildOrder(ta *entity.FcTransfersApply) (*transfer.XlmOrderReq, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.Find(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, err = decimal.NewFromString(toAddrs[0].ToAmount)
	if err != nil {
		return nil, err
	}
	coinType := strings.ToLower(ta.CoinName)
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	coin := global.CoinDecimal[coinType]
	if coin == nil {
		return nil, fmt.Errorf("读取 %s coinSet 设置异常", coinType)
	}
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount >= ? and forzen_amount = 0", coinType, toAddrs[0].ToAmount).
		And(builder.In("address", coldAddrs[0].Address)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址\n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]

	//填充参数
	orderReq := &transfer.XlmOrderReq{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName
	orderReq.Worker = service.GetWorker(ta.CoinName)
	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Memo = ta.Memo
	orderReq.Token = ta.Eostoken
	//将金额转换为int64
	orderReq.Amount = toAmount.Shift(int32(coin.Decimal)).IntPart()
	return orderReq, nil
}

//===========================================valid address=============================================//

var crc16tab = [256]uint16{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
	0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
	0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
	0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
	0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
	0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
	0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
	0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
	0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
	0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
	0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
	0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
	0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
	0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
	0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
	0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
	0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
	0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
	0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
	0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
	0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
	0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
	0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
	0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
	0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
	0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
	0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
	0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
	0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
	0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0,
}

var decodingTable = initDecodingTable()

var ErrInvalidVersionByte = errors.New("invalid version byte")

// VersionByte represents one of the possible prefix values for a StrKey base
// string--the string the when encoded using base32 yields a final StrKey.
type VersionByte byte

const (
	//VersionByteAccountID is the version byte used for encoded stellar addresses
	VersionByteAccountID VersionByte = 6 << 3 // Base32-encodes to 'G...'

	//VersionByteAccountID is the version byte used for encoded stellar multiplexed addresses
	VersionByteMuxedAccount = 12 << 3 // Base32-encodes to 'M...'
)

func initDecodingTable() [256]byte {
	var localDecodingTable [256]byte
	for i := range localDecodingTable {
		localDecodingTable[i] = 0xff
	}
	for i, ch := range []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567") {
		localDecodingTable[ch] = byte(i)
	}
	return localDecodingTable
}

func decode(expected VersionByte, src string) ([]byte, error) {
	if err := checkValidVersionByte(expected); err != nil {
		return nil, err
	}

	raw, err := decodeString(src)
	if err != nil {
		return nil, err
	}

	// decode into components
	version := VersionByte(raw[0])
	vp := raw[0 : len(raw)-2]
	payload := raw[1 : len(raw)-2]
	checksum := raw[len(raw)-2:]

	// ensure version byte is expected
	if version != expected {
		return nil, ErrInvalidVersionByte
	}

	// ensure checksum is valid
	if err := Validate(vp, checksum); err != nil {
		return nil, err
	}

	// if we made it through the gaunlet, return the decoded value
	return payload, nil
}

// checkValidVersionByte returns an error if the provided value
// is not one of the defined valid version byte constants.
func checkValidVersionByte(version VersionByte) error {
	if version == VersionByteAccountID {
		return nil
	}

	if version == VersionByteMuxedAccount {
		return nil
	}

	return ErrInvalidVersionByte
}

// decodeString decodes a base32 string into the raw bytes, and ensures it could
// potentially be strkey encoded (i.e. it has both a version byte and a
// checksum, neither of which are explicitly checked by this func)
func decodeString(src string) ([]byte, error) {
	// operations on strings are expensive since it involves unicode parsing
	// so, we use bytes from the beginning
	srcBytes := []byte(src)
	// The minimal binary decoded length is 3 bytes (version byte and 2-byte CRC) which,
	// in unpadded base32 (since each character provides 5 bits) corresponds to ceiling(8*3/5) = 5
	if len(srcBytes) < 5 {
		return nil, fmt.Errorf("strkey is %d bytes long; minimum valid length is 5", len(srcBytes))
	}
	// SEP23 enforces strkeys to be in canonical base32 representation.
	// Go's decoder doesn't help us there, so we need to do it ourselves.
	// 1. Make sure there is no full unused leftover byte at the end
	//   (i.e. there shouldn't be 5 or more leftover bits)
	leftoverBits := (len(srcBytes) * 5) % 8
	if leftoverBits >= 5 {
		return nil, errors.New("non-canonical strkey; unused leftover character")
	}
	// 2. In the last byte of the strkey there may be leftover bits (4 at most, otherwise it would be a full byte,
	//    which we have for checked above). If there are any leftover bits, they should be set to 0
	if leftoverBits > 0 {
		lastChar := srcBytes[len(srcBytes)-1]
		decodedLastChar := decodingTable[lastChar]
		leftoverBitsMask := byte(0x0f) >> (4 - leftoverBits)
		if decodedLastChar&leftoverBitsMask != 0 {
			return nil, errors.New("non-canonical strkey; unused bits should be set to 0")
		}
	}
	n, err := base32.StdEncoding.WithPadding(base32.NoPadding).Decode(srcBytes, srcBytes)
	if err != nil {
		return nil, fmt.Errorf("base32 decode failed,err=%v", err)
	}

	return srcBytes[:n], nil
}

// Validate returns an error if the provided checksum does not match
// the calculated checksum of the provided data
func Validate(data []byte, expected []byte) error {

	actual := Checksum(data)

	// validate the provided checksum against the calculated
	if !bytes.Equal(actual, expected) {
		return errors.New("invalid checksum")
	}

	return nil
}

// Checksum returns the 2-byte checksum for the provided data
func Checksum(data []byte) []byte {
	var crc uint16
	var out bytes.Buffer
	for _, b := range data {
		crc = ((crc << 8) & 0xffff) ^ crc16tab[((crc>>8)^uint16(b))&0x00FF]
	}

	err := binary.Write(&out, binary.LittleEndian, crc)
	if err != nil {
		panic(err)
	}

	return out.Bytes()
}

func (srv *XlmTransferService) buildOrderHot(ta *entity.FcTransfersApply, coinDecimal int32) (*transfer.XlmOrderHotRequest, error) {
	var (
		fromAddr string
		toAddr   string
		toAmount decimal.Decimal
	)
	// 查找from地址和金额
	coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
		"type":        address.AddressTypeCold,
		"status":      address.AddressStatusAlloc,
		"platform_id": ta.AppId,
		"coin_name":   ta.CoinName,
	})
	if err != nil {
		return nil, err
	}
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddr = toAddrs[0].Address
	toAmount, err = decimal.NewFromString(toAddrs[0].ToAmount)
	if err != nil {
		return nil, err
	}
	coinType := ta.CoinName
	if ta.Eoskey != "" {
		coinType = strings.ToLower(ta.Eoskey)
	}
	//
	fee, err := decimal.NewFromString(ta.Fee)
	if err != nil {
		return nil, fmt.Errorf("order ")
	}
	fromAmount := toAmount.Add(fee)
	fromAddrs, err := entity.FcAddressAmount{}.FindAddress(builder.Expr("coin_type = ? and amount > ? and forzen_amount = 0", coinType, fromAmount.String()).
		And(builder.In("address", coldAddrs)), 0)
	if err != nil {
		return nil, fmt.Errorf("err:%s", err.Error())
	}
	if len(fromAddrs) == 0 {
		return nil, fmt.Errorf("outorderNo:%s 没有符合条件的出账地址，大于0.004 \n amount: %v \n to: %s \n ", ta.OutOrderid, toAddrs[0].ToAmount, toAddr)
	}
	fromAddr = fromAddrs[0]

	orderReq := &transfer.XlmOrderHotRequest{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchId = int64(ta.AppId)
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = ta.CoinName

	orderReq.FromAddress = fromAddr
	orderReq.ToAddress = toAddr
	orderReq.Amount = toAmount.Shift(coinDecimal).String()
	orderReq.Memo = ta.Memo
	orderReq.Fee = fee.Shift(coinDecimal).String()
	orderReq.Token = ta.Eostoken
	return orderReq, nil
}

func (srv *XlmTransferService) walletServerCreateHot(orderReq *transfer.XlmOrderHotRequest) (string, error) {
	cfg, ok := conf.Cfg.HotServers[srv.CoinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", srv.CoinName)
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", srv.CoinName, string(dd))
	data, err := util.PostJsonByAuth(fmt.Sprintf("%s/v1/%s/transfer", cfg.Url, srv.CoinName), cfg.User, cfg.Password, orderReq)
	if err != nil {
		return "", err
	}
	log.Infof("%s 交易返回内容 :%s", srv.CoinName, string(data))
	result, err := transfer.DecodeTransferHotResp(data)

	if err != nil || result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s,err: %v", orderReq.OuterOrderNo, err)
	}
	if result.Code != 0 {
		log.Error(result)
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，outOrderId：%s，err:%s", orderReq.OuterOrderNo, string(data))
	}
	return result.Txid, nil
}
