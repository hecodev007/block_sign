package zecservice

import (
	"encoding/csv"
	"fmt"
	"github.com/group-coldwallet/zecserver/model"
	"github.com/group-coldwallet/zecserver/model/bo"
	"github.com/group-coldwallet/zecserver/model/global"
	"github.com/group-coldwallet/zecserver/model/vo"
	"github.com/group-coldwallet/zecserver/util"
	"github.com/group-coldwallet/zecserver/util/rylink"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

type ZecServiceV1 struct {
}

//源码签名
func NewZecServiceV1() BasicService {
	return &ZecServiceV1{}
}

func (zec *ZecServiceV1) SignTx(tpl *bo.ZecTxTpl) (hex string, err error) {

	var (
		fromAmount int64
		toAmount   int64
		fee        int64
	)

	//简单校验参数
	if len(tpl.TxIns) == 0 || len(tpl.TxOuts) == 0 {
		err = fmt.Errorf("empty datas")
		logrus.Error(err.Error())
		return "", err
	}
	//txin校验
	for i, v := range tpl.TxIns {
		if v.FromAddr == "" {
			err = fmt.Errorf("index:%d,empty address", i)
			logrus.Error(err.Error())
			return "", err
		}
		if v.FromIndex < 0 {
			err = fmt.Errorf("index:%d, address:%s, error vout", i, v.FromAddr)
			logrus.Errorf(err.Error())

			return "", err
		}
		if v.FromAmount < 0 {
			err = fmt.Errorf("index:%d, address:%s, error amount ", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}

		if v.FromTxid == "" {
			err = fmt.Errorf("index:%d, address:%s, error txid", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		//查询私钥
		privkey, _ := global.GetValue(v.FromAddr)
		if privkey == "" {
			err = fmt.Errorf("index:%d, address:%s, miss privkey", i, v.FromAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		//添加私钥
		tpl.TxIns[i].FromPrivkey = privkey

		fromAmount += v.FromAmount

	}

	//txout校验
	for i, v := range tpl.TxOuts {
		if v.ToAddr == "" {
			err = fmt.Errorf("index:%d,empty toAddress", i)
			logrus.Error(err.Error())
			return "", err
		}
		if v.ToAmount < 546 {
			err = fmt.Errorf("index:%d, address:%s, error toAmount，min value:546", i, v.ToAddr)
			logrus.Errorf(err.Error())
			return "", err
		}
		toAmount += v.ToAmount
	}

	if fromAmount < toAmount {
		err = fmt.Errorf("Wrong output total,fromAmount:%d,toAmount:%d", fromAmount, toAmount)
		logrus.Errorf(err.Error())
		return "", err
	}

	if fromAmount == toAmount {
		err = fmt.Errorf("Must pay the fee,fromAmount:%d,toAmount:%d", fromAmount, toAmount)
		logrus.Errorf(err.Error())
		return "", err
	}

	fee = fromAmount - toAmount
	//下限0.00001
	if fee < 1000 {
		err = fmt.Errorf("The fee is too low, fee:%d", fee)
		logrus.Errorf(err.Error())
		return "", err
	}

	//上限 1
	if fee > 100000000 {
		err = fmt.Errorf("The fee is too high, fee:%d", fee)
		logrus.Errorf(err.Error())
		return "", err
	}
	hex, err = zec.zecSignTxTpl(tpl)
	if err != nil {
		logrus.Error(err.Error())
		return "", err
	}
	return hex, nil
}

func (zec *ZecServiceV1) zecSignTxTpl(tpl *bo.ZecTxTpl) (hex string, err error) {

	vins := make([]*rylink.RpcCreatetx, 0)
	prevtxs := make([]rylink.SigntxPrevtxs, 0)
	privatekeys := make([]string, 0)
	for _, v := range tpl.TxIns {
		amount, _ := decimal.New(v.FromAmount, -8).Float64()
		vins = append(vins, &rylink.RpcCreatetx{
			Txid: v.FromTxid,
			Vout: int(v.FromIndex),
		})
		prevtxs = append(prevtxs, rylink.SigntxPrevtxs{
			Txid:         v.FromTxid,
			Vout:         int(v.FromIndex),
			ScriptPubKey: v.FromScriptPubKey,
			Amount:       amount,
			RedeemScript: v.FromRedeemScript,
		})
		privatekeys = append(privatekeys, v.FromPrivkey)
	}

	vouts := make(map[string]float64)
	for _, v := range tpl.TxOuts {
		amount, _ := decimal.New(v.ToAmount, -8).Float64()
		if amount > 0 {
			vouts[v.ToAddr] = amount
		}
	}

	//结果进行签名
	signResult, err := rylink.SignTxTpl(tpl)
	if err != nil {
		return "", err
	}
	return signResult, nil

}

func (zec *ZecServiceV1) CreateAddr(params *bo.CreateAddrParam, createPath, readPath string) (*vo.CreateAddrResult, error) {
	createPath = createPath + "/" + params.MchId
	readPath = readPath + "/" + params.MchId
	result := &vo.CreateAddrResult{
		Num: params.Num,
		MchInfo: model.MchInfo{
			OrderId:  params.OrderId,
			MchId:    params.MchId,
			CoinName: params.CoinName,
		},
	}

	//临时存储地址
	resultAddrs := make([]string, 0)
	//先批量生成完成，再写入文件
	addrs := make([]*bo.AddressInfo, 0)
	filename := createPath + "/" + params.CoinName + "_%s_usb_" + params.OrderId + ".csv"
	fileAPath := fmt.Sprintf(filename, "a")
	fileBPath := fmt.Sprintf(filename, "b")
	fileCPath := fmt.Sprintf(filename, "c")
	fileDPath := fmt.Sprintf(filename, "d")

	//A文件判断,
	_, err := os.Stat(fileAPath)
	if err == nil {
		log.Println("已经存在a文件 ", err)
		//读取A文件，直接读取地址
		resultAddrs, err := util.ReadCsv(fileAPath, 0)
		if err != nil {
			log.Println("read fileA error ", err)
			return nil, err
		}
		result.Addrs = resultAddrs
		return result, nil
	}
	//创建多层级目录
	_, err = util.CreateDirAll(createPath + "/")
	if err != nil {
		log.Println("create create dir error ", err)
		return nil, err
	}
	//创建多层级目录(备份目录，只存储AB文件)
	_, err = util.CreateDirAll(readPath + "/")
	if err != nil {
		log.Println("create copy dir error ", err)
		return nil, err
	}
	//生成地址
	for i := 0; i < params.Num; i++ {
		zaddr, zprivkey, err := rylink.CreateAddress()
		if err != nil {
			log.Printf("createn new address error,numbers: %d, error:%v", i, err)
			return nil, fmt.Errorf("createn new address error,numbers: %d, error:%v", i, err)
		}
		addrs = append(addrs, &bo.AddressInfo{
			Address:    zaddr,
			PrivateKey: zprivkey,
		})
	}
	if len(addrs) != int(params.Num) {
		log.Println(fmt.Sprintf("createn address error,len :%d", len(addrs)))
		return nil, fmt.Errorf("createn address error,len :%d", len(addrs))
	}
	fileA, err := os.Create(fileAPath)
	if err != nil {
		return nil, err
	}
	defer fileA.Close()

	fileB, err := os.Create(fileBPath)
	if err != nil {
		return nil, err
	}
	defer fileB.Close()

	//C文件
	fileC, err := os.Create(fileCPath)
	if err != nil {
		return nil, err
	}
	defer fileC.Close()

	//D文件
	_, err = os.Stat(fileDPath)
	fileD, err := os.Create(fileDPath)
	defer fileD.Close()

	wa := csv.NewWriter(fileA) //创建一个新的写入文件流
	wb := csv.NewWriter(fileB)
	wc := csv.NewWriter(fileC)
	wd := csv.NewWriter(fileD)
	for _, info := range addrs {
		aesKey := util.RandBase64Key()
		aesPrivateKey, err := util.AesBase64Crypt([]byte(info.PrivateKey), aesKey, true)
		if err != nil {
			log.Println("util.AesBase64Crypt error: ", err)
			continue
		}
		wa.Write([]string{info.Address, string(aesPrivateKey)})
		wb.Write([]string{info.Address, string(aesKey)})
		wc.Write([]string{info.Address, string(info.PrivateKey)})
		wd.Write([]string{info.Address})
		resultAddrs = append(resultAddrs, info.Address)
	}
	wa.Flush()
	wb.Flush()
	wc.Flush()
	wd.Flush()
	log.Println(fileAPath, fileBPath, fileCPath, fileDPath)
	result.Addrs = resultAddrs
	//生成完成，ab文件异步复制到读取目录
	//复制文件到readPath

	copyName := readPath + "/" + params.CoinName + "_%s_usb_" + params.OrderId + ".csv"
	copyA := fmt.Sprintf(copyName, "a")
	copyB := fmt.Sprintf(copyName, "b")
	a, err := util.FileCopy(fileAPath, copyA)
	if err != nil {
		log.Println(fmt.Errorf("copy a error:%v", err.Error()))
	} else {
		log.Println(fmt.Errorf("copy a success:%d", a))
		b, err := util.FileCopy(fileBPath, copyB)
		if err == nil {
			log.Println(fmt.Errorf("copy b success:%d", b))
		}

	}
	return result, nil
}
