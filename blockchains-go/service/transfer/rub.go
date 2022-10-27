package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"strconv"
	"strings"

	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/shopspring/decimal"

	"xorm.io/builder"
)

type RubTransferService struct {
	CoinName string
}

func NewRubTransferService() service.TransferService {
	return &RubTransferService{CoinName: "rub"}
}
func (srv *RubTransferService) TransferHot(ta *entity.FcTransfersApply) (txid string, err error) {
	var (
		orderReq *transfer.RubOrderReq
	)
	orderReq, err = srv.buildHotOrder(ta)
	if err != nil {
		log.Errorf("下单表订单id：%d,构建异常:%s", ta.Id, err.Error())
		return "", err
	}
	return srv.walletServerCreateHot(orderReq)
}
func (srv *RubTransferService) TransferCold(ta *entity.FcTransfersApply) error {
	return errors.New("this is a hot wallet")
}
func (srv *RubTransferService) VaildAddr(address string) error {
	return RubValidAddress(address)
}

func (srv *RubTransferService) walletServerCreateHot(orderReq *transfer.RubOrderReq) (string, error) {

	var (
		coinName string
		data     []byte
		err      error
	)

	coinName = strings.ToLower(orderReq.CoinName)
	cfg, ok := conf.Cfg.HotServers[coinName]
	if !ok {
		return "", fmt.Errorf("don't find %s config", coinName)
	}
	log.Infof("rubchain 执行币种：%s", coinName)
	switch coinName {
	case "rub":
		data, err = util.PostJsonByAuth(fmt.Sprintf("%s/v1/rubychain/transfer", cfg.Url), cfg.User, cfg.Password, orderReq)
	case "wbc":
		data, err = util.PostJsonByAuth(fmt.Sprintf("%s/v1/ruby-wbc/transfer", cfg.Url), cfg.User, cfg.Password, orderReq)
	default:
		return "", fmt.Errorf("rub链不存在相关币种[%s]设置", coinName)
	}
	if err != nil {
		return "", err
	}
	dd, _ := json.Marshal(orderReq)
	log.Infof("%s 交易发送内容 :%s", srv.CoinName, string(dd))
	log.Infof("%s 交易返回内容 :%s", srv.CoinName, string(data))
	result := transfer.DecodeWalletServerRespOrder(data)
	if result == nil {
		return "", fmt.Errorf("order表 请求下单接口失败，outOrderId：%s", orderReq.OuterOrderNo)
	}
	if result.Code != 0 || result.Data == nil {
		return "", fmt.Errorf("order表 请求下单接口返回值失败,服务器返回异常，data:%s,outOrderId：%s", string(data), orderReq.OuterOrderNo)
	}
	txid := result.Data.(string)
	return txid, nil
}

func (srv *RubTransferService) buildHotOrder(ta *entity.FcTransfersApply) (*transfer.RubOrderReq, error) {
	var (
		changeAddress string
		toAddress     string
		fee           int64           = 0 //默认等于0
		toAmount      decimal.Decimal     //发送金额
		coinName      string              //币种名字
	)
	if ta.Eoskey != "" {
		coinName = ta.Eoskey
	} else {
		coinName = ta.CoinName
	}

	//查询找零地址，需要查询主链币种找零地址
	changes, err := dao.FcGenerateAddressListFindChangeAddr(ta.AppId, ta.CoinName)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("rub 商户=[%d],查询rub找零地址失败", ta.AppId)
	}
	//随机选择
	randIndex := util.RandInt64(0, int64(len(changes)))
	changeAddress = changes[randIndex]
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": ta.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", ta.Id, ta.OutOrderid)
	}
	toAddress = toAddrs[0].Address
	toAmount, _ = decimal.NewFromString(toAddrs[0].ToAmount)
	if toAmount.IsZero() {
		return nil, errors.New("rub toAmount  is zero")
	}
	coinSet := global.CoinDecimal[coinName]
	if coinSet == nil {
		return nil, fmt.Errorf("缺少币种信息")
	}
	tos := transfer.RubyChainParamsTransferTos{
		ToAddress: toAddress,
		ToAmount:  toAmount.Shift(int32(coinSet.Decimal)).IntPart(),
	}
	var toses []transfer.RubyChainParamsTransferTos
	toses = append(toses, tos)
	orderReq := &transfer.RubOrderReq{}
	orderReq.ApplyId = int64(ta.Id)
	orderReq.OuterOrderNo = ta.OutOrderid
	orderReq.OrderNo = ta.OrderId
	orderReq.MchName = ta.Applicant
	orderReq.CoinName = coinName
	//orderReq.Worker = service.GetWorker(srv.CoinName)
	orderReq.ChangeAddress = changeAddress
	orderReq.Fee = fee
	orderReq.Tos = toses

	return orderReq, nil
}

//===============================================验证地址=====================================================//

const (
	PubKeyHashVersion = "007f5512"
	PrivateKeyVersion = "8028effe"
	ChecksumValue     = "c81bd898"
)

var (
	hexArray = map[uint8]string{
		48:  "0000",
		49:  "0001",
		50:  "0010",
		51:  "0011",
		52:  "0100",
		53:  "0101",
		54:  "0110",
		55:  "0111",
		56:  "1000",
		57:  "1001",
		97:  "1010",
		98:  "1011",
		99:  "1100",
		100: "1101",
		101: "1110",
		102: "1111",
	}
)

/*
reference: github.com/RubyChainNet/sdk/rubychainjs-lib/static/common/rubyjs-lib-release1.0.js
*/
func RubValidAddress(address string) error {
	if !strings.HasPrefix(address, "1") {
		return errors.New("地址不是以1开头")
	}
	adBytes := base58.Decode(address)
	if adBytes == nil || len(adBytes) < 10 {
		return errors.New("base58 decode address error")
	}
	adHex := hex.EncodeToString(adBytes)

	xorv := adHex[len(adHex)-8:]
	rubPubHash := adHex[:len(adHex)-8]
	err := checkPubHashVersion(rubPubHash)
	if err != nil {
		return err
	}
	checksum1 := xor(ChecksumValue, xorv)
	doubleSha := doubleSha256(rubPubHash)
	checksum2 := doubleSha[:8]
	if strings.Compare(checksum1, checksum2) != 0 {
		return errors.New("invalid checksum")
	}
	return nil
}

/*
接收两个16进制字符串
*/
func xor(a, b string) string {
	as := hexToBin(a)
	bs := hexToBin(b)
	var buf string
	for i := 0; i < len(as); i++ {
		m := []byte(as[i : i+1])[0]
		n := []byte(bs[i : i+1])[0]
		s := m ^ n
		buf = buf + strconv.FormatUint(uint64(s), 10)
	}
	return binToHex(buf)
}
func checkPubHashVersion(rubyPubHash string) error {
	if len(rubyPubHash) != 48 {
		return errors.New("ruby public hash length is not equal 48")
	}
	phv := ""
	for i := 0; i < len(rubyPubHash); i += 12 {
		phv = phv + rubyPubHash[i:i+2]
	}
	if strings.Compare(phv, PubKeyHashVersion) != 0 {
		return errors.New("invalid ruby public hash")
	}
	return nil
}

/*
16进制转换为二进制
*/
func hexToBin(data string) string {
	bin := ""
	for i := 0; i < len(data); i++ {
		for k, v := range hexArray {
			if data[i] == k {
				bin = bin + v
			}
		}
	}
	return bin
}

/*
二进制转换为16进制
*/
func binToHex(data string) string {
	hex := ""
	if len(data) < 4 {
		return hex
	}
	if (len(data) % 4) != 0 {
		b := data[:len(data)%4]
		hex = b + data
		return hex
	}
	for i := 0; i < len(data); i += 4 {
		for k, v := range hexArray {
			if data[i:i+4] == v {
				hex = hex + string(k)
			}
		}
	}
	return hex
}

func doubleSha256(data string) string {
	d, _ := hex.DecodeString(data)
	sha1 := sha256.Sum256(d)
	sha2 := sha256.Sum256(sha1[:])
	return hex.EncodeToString(sha2[:])
}
