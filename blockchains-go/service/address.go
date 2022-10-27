package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/db"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxApplyAddrNum = 500
)

func GetColdAddre(appId int, coinName string) ([]*entity.FcGenerateAddressList, error) {
	return dao.FcGenerateAddressListFindAddressesData(1, 2, appId, coinName)
}

//分配地址
func AssignMchAddrs(applyAddrReq model.ApplyAddrReq) ([]*entity.FcGenerateAddressList, error) {
	var (
		err      error
		mchName  string
		mvi      dao.MchVerifyInfo
		signData map[string]string
		ok       bool
		as       []*entity.FcGenerateAddressList
	)

	applyAddrReq.Sfrom = strings.Trim(applyAddrReq.Sfrom, " ")
	if mchName = applyAddrReq.Sfrom; len(mchName) == 0 {
		return nil, errors.New("sfrom is empty")
	}

	applyAddrReq.CoinName = strings.Trim(applyAddrReq.CoinName, " ")
	if len(applyAddrReq.CoinName) == 0 {
		return nil, errors.New("coinName is empty")
	}

	applyAddrReq.Sign = strings.Trim(applyAddrReq.Sign, " ")
	if len(applyAddrReq.Sign) == 0 {
		return nil, errors.New("sign is empty")
	}

	applyAddrReq.OutOrderId = strings.Trim(applyAddrReq.OutOrderId, " ")
	if len(applyAddrReq.OutOrderId) == 0 {
		return nil, errors.New("outOrderId is empty")
	}
	if applyAddrReq.Num <= 0 || applyAddrReq.Num > maxApplyAddrNum {
		return nil, errors.New(fmt.Sprintf("num must between 1 -- %d", maxApplyAddrNum))
	}

	signData = map[string]string{
		"sign":       applyAddrReq.Sign,
		"sfrom":      applyAddrReq.Sfrom,
		"outOrderId": applyAddrReq.OutOrderId,
		"coinName":   applyAddrReq.CoinName,
		"num":        strconv.FormatInt(applyAddrReq.Num, 10),
	}
	if ok, err = VerifySign(mchName, signData); !ok {
		//log写错误原因
		return nil, fmt.Errorf("signature no pass: %w", err)
	}

	if mvi, err = dao.GetMchVerifyInfo(applyAddrReq.Sfrom); err != nil {
		//log写错误原因
		return nil, errors.New("get mch info exception")
	}

	applyAddrReq.CoinName = strings.ToLower(applyAddrReq.CoinName)
	if as, _, err = dao.AssignMchAddrs(mvi.MchId, "", applyAddrReq.CoinName, applyAddrReq.OutOrderId, applyAddrReq.Num); err != nil {
		//log写错误原因
		return nil, errors.New("assign mch address exception")
	}
	return as, nil
}

//分配地址
func AssignMchAddrsV2(applyAddrReq model.ApplyAddrReq) ([]*entity.FcGenerateAddressList, error) {
	var (
		err error
		as  []*entity.FcGenerateAddressList
	)
	if applyAddrReq.CoinName == "" {
		return nil, errors.New("coin_name is empty")
	}
	if applyAddrReq.OutOrderId == "" {
		return nil, errors.New("out_order_id is empty")
	}
	if applyAddrReq.Num <= 0 || applyAddrReq.Num > maxApplyAddrNum {
		return nil, errors.New(fmt.Sprintf("num must between 1 -- %d", maxApplyAddrNum))
	}

	//mch, err := dao.FcMchFindByApikey(applyAddrReq.ClientId)
	//if err != nil {
	//	return nil, fmt.Errorf("get mch info exception:%s", err.Error())
	//}
	mch, ok := global.MchBaseInfo[applyAddrReq.ClientId]
	if !ok {
		global.ReloadMchBaseInfo()
		mch, ok = global.MchBaseInfo[applyAddrReq.ClientId]
		if !ok {
			return nil, fmt.Errorf("get mch info error")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("get mch info exception:%s", err.Error())
	}

	var result *entity.FcGenerateAddressList
	result, err = dao.FcGenerateAddressFindByOutOrderNo(applyAddrReq.OutOrderId, mch.AppId)
	if result != nil || err.Error() != "Not Fount!" {
		if err != nil {
			log.Errorf("FcGenerateAddressFindByOutOrderNo 异常：%s", err.Error())
		}
		log.Errorf("订单号重复：%s", applyAddrReq.OutOrderId)
		return nil, fmt.Errorf("订单号重复：%s", applyAddrReq.OutOrderId)
	}
	//数据库内部是小写。但是目前线上是大小写掺杂
	applyAddrReq.CoinName = strings.ToLower(applyAddrReq.CoinName)
	var remain int
	if as, remain, err = dao.AssignMchAddrs(int64(mch.AppId), mch.MchName, applyAddrReq.CoinName, applyAddrReq.OutOrderId, applyAddrReq.Num); err != nil {
		//log写错误原因
		log.Errorf("分配地址异常：%s", err.Error())
		if strings.Contains(err.Error(), "no enough") {
			go CreateAssignMchAddr(int64(mch.AppId), mch.MchName, applyAddrReq.CoinName, int(applyAddrReq.Num), remain)
			if remain < 0 {
				return nil, fmt.Errorf("Creating,Try again in a minute")
			}
		}
		return nil, errors.New("assign mch address exception")
	} else {
		log.Errorf("分配地址成功：as %+v", as)
		log.Errorf("分配地址成功：remain %+v", remain)
		go CreateAssignMchAddr(int64(mch.AppId), mch.MchName, applyAddrReq.CoinName, int(applyAddrReq.Num), remain)
	}
	return as, nil
}

//自动创建地址
func CreateAssignMchAddr(mchId int64, mchName, coinName string, num, remain int) {
	min := 10
	max := 1000
	var creatNum int
	if remain <= 0 {
		if (2 * num) < min {
			creatNum = min
		} else if (2 * num) > max {
			if num < max {
				creatNum = max
			} else {
				creatNum = num
			}
		} else {
			creatNum = 2 * num
		}
	} else {
		if remain < num {
			if num < min {
				creatNum = min
			} else {
				creatNum = num
			}
		}
	}
	if creatNum <= 0 {
		return
	}
	//创建地址
	req := model.GenerateTransportKeyHandleRequest{
		CoinCode:       coinName,
		Mch:            mchName,
		Count:          int64(creatNum),
		//Type:           model.TypeGtkGenerateAddr,
		//AssignSignerNo: "tiger00001",
		//SignerList:     []string{"lion00001"},
	}
	pByte, _ := json.Marshal(req)
	param := string(pByte)

	walleType := global.WalletType(coinName, int(mchId))
	if walleType == status.WalletType_Cold {
		log.Infof("冷钱包创建地址 %v",coinName)
		log.Infof("url=%v,param=%v",conf.Cfg.Commandcenter.Url,param)
		resp, err := http.Post(conf.Cfg.Commandcenter.Url, "application/json", strings.NewReader(param))
		if err != nil {
			log.Errorf("自动创建地址错误 req:%+v,\n err：%s\n", req, err.Error())
			return
		}
		defer resp.Body.Close()
		result, _ := ioutil.ReadAll(resp.Body)
		var f model.BaseResult
		err = json.Unmarshal(result, &f)
		if err != nil {
			log.Errorf("自动创建地址 Unmarshal 错误 result:%+v,\n err：%s\n", string(result), err.Error())
			return
		}
		log.Infof("冷钱包创建地址 result =%+v",f)
		if f.Code != 0 {
			log.Errorf("自动创建地址 Code 错误 BaseResult:%+v,\n err：%s\n", f, err.Error())
			return
		}
		//成功保存到redis
		key := fmt.Sprintf("custody:create-address:%v", mchName)
		redis.Client.Set(key, mchId, 0*time.Second)
	} else { //热钱包
		log.Infof("热钱包创建地址 %v",coinName)
		cfg, ok := conf.Cfg.HotServers[coinName]
		if !ok {
			err := fmt.Errorf("don't find %s config", coinName)
			log.Errorf("热钱包创建地址 err %v", err)
			return
		}
		url := fmt.Sprintf("%s/v1/%s/createAddr", cfg.Url, coinName)
		log.Infof("CreateAssignMchAddr 热钱包创建地址 url %v\n", url)
		log.Infof("CreateAssignMchAddr 热钱包创建地址 param %+v\n", param)
		resp, err := http.Post(url, "application/json", strings.NewReader(param))
		if err != nil {
			log.Errorf("自动创建地址错误 req:%+v,\n err：%s\n", req, err.Error())
			return
		}
		defer resp.Body.Close()
		result, _ := ioutil.ReadAll(resp.Body)
		var f model.BaseResult
		err = json.Unmarshal(result, &f)
		if err != nil {
			log.Errorf("热钱包创建地址 Unmarshal 错误 result:%+v,\n err：%s\n", string(result), err.Error())
			return
		}
		if f.Code != 0 {
			log.Errorf("热钱包创建地址 Code 错误 BaseResult:%+v\n", f)
			return
		}
		log.Infof("热钱包创建地址 BaseResult:%+v,\n", f)

		if f.Data != nil {
			resByte, resByteErr := json.Marshal(f.Data)
			if resByteErr != nil {
				err = fmt.Errorf("json Marshal err:%v", resByteErr.Error())
				return
			}
			address := model.AddressRsp{}
			err = json.Unmarshal(resByte, &address)
			if err != nil {
				err = fmt.Errorf("json Unmarshal err:%v", err.Error())
				return
			}
			//成功保存到redis
			key := fmt.Sprintf("custody:create-address:%v", mchName)
			redis.Client.Set(key, mchId, 0*time.Second)
			SaveBackAddress(address)
		}
	}
	return
}

//创建地址回传
func SaveBackAddress(rsp model.AddressRsp) (err error) {

	log.Infof("SaveBackAddress rsp:%+v,\n", SaveBackAddress)
	key := fmt.Sprintf("custody:create-address:%v", rsp.Mch)
	mchIdStr, _ := redis.Client.Get(key)
	mchId, _ := strconv.ParseInt(mchIdStr, 10, 64)
	if mchIdStr == "" {
		mch, _ := dao.FcMchFindByPlatform(rsp.Mch)
		if mch == nil || mch.Id == 0 {
			err = errors.New("商户不存在")
			return
		}
		mchId = int64(mch.Id)
	}

	coinInfo, _ := dao.FcCoinSetGetByName(rsp.CoinCode, 1)
	if coinInfo.Id == 0 {
		err = fmt.Errorf("币种(%v)不存在", rsp.CoinCode)
		return
	}

	//获取已有的
	var (
		one   int
		three int
	)
	one, three, err = dao.GetMchAddr2(mchId, rsp.CoinCode)
	log.Infof("SaveBackAddress one,three :%+v,:%+v,\n", one, three)
	if err != nil {
		return
	}
	tx := db.Conn.NewSession()
	err = tx.Begin()
	if err = tx.Begin(); err != nil {
		return fmt.Errorf("db begin error: %w", err)
	}
	addArr := rsp.Address
	t := time.Now()
	tNum := t.Unix()
	syncArr := make([]string, 0)
	if one <= 0 {
		var oneAddress string
		if len(addArr) >= 1 {
			oneAddress = addArr[0]
			oneItem := entity.FcGenerateBeforeAddressList{
				PlatformId: int(mchId),
				CoinId:     coinInfo.Id,
				CoinName:   rsp.CoinCode,
				Address:    oneAddress,
				Status:     2,
				Type:       1,
				OutOrderid: rsp.BatchNo,
				Createtime: int(tNum),
				Lastmodify: t,
			}
			err = dao.InsertMchFirstAddress(&oneItem, tx)
			log.Infof("SaveBackAddress InsertMchFirstAddress err :%+v,\n", err)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("db error: %w", err)
			}
			oneFAL := entity.FcGenerateAddressList{
				PlatformId: int(mchId),
				CoinId:     coinInfo.Id,
				CoinName:   rsp.CoinCode,
				Address:    oneAddress,
				Status:     2,
				Type:       1,
				OutOrderid: rsp.BatchNo,
				Createtime: int(tNum),
				Lastmodify: t,
			}
			err = dao.InsertMchFirstGALAddress(&oneFAL, tx)
			log.Infof("SaveBackAddress InsertMchFirstGALAddress err :%+v,\n", err)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("db error: %w", err)
			}
			syncArr = append(syncArr, oneAddress)
			addArr = addArr[1:]
		}
	}
	if three <= 0 {
		var oneAddress string
		if len(addArr) >= 1 {
			oneAddress = addArr[0]
			oneItem := entity.FcGenerateBeforeAddressList{
				PlatformId: int(mchId),
				CoinId:     coinInfo.Id,
				CoinName:   rsp.CoinCode,
				Address:    oneAddress,
				Status:     2,
				Type:       3,
				OutOrderid: rsp.BatchNo,
				Createtime: int(tNum),
				Lastmodify: t,
			}
			err = dao.InsertMchFirstAddress(&oneItem, tx)
			log.Infof("SaveBackAddress InsertMchFirstAddress3 err :%+v,\n", err)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("db error: %w", err)
			}
			oneFAL := entity.FcGenerateAddressList{
				PlatformId: int(mchId),
				CoinId:     coinInfo.Id,
				CoinName:   rsp.CoinCode,
				Address:    oneAddress,
				Status:     2,
				Type:       3,
				OutOrderid: rsp.BatchNo,
				Createtime: int(tNum),
				Lastmodify: t,
			}
			err = dao.InsertMchFirstGALAddress(&oneFAL, tx)
			log.Infof("SaveBackAddress InsertMchFirstGALAddress3 err :%+v,\n", err)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("db error: %w", err)
			}
			syncArr = append(syncArr, oneAddress)
			addArr = addArr[1:]
		}
	}
	adds := make([]string, 0)
	for _, item := range addArr {
		adds = append(adds, item)
		if len(adds) == 50 {
			err = dao.InsertBatchMchAddress(int(mchId), coinInfo.Id, rsp.CoinCode, rsp.BatchNo, adds, t, tx)
			log.Infof("SaveBackAddress InsertMchFirstAddress50 err :%+v,\n", err)
			if err != nil {
				tx.Rollback()
				return
			} else {
				adds = make([]string, 0)
			}
		}
	}
	if len(adds) > 0 {
		err = dao.InsertBatchMchAddress(int(mchId), coinInfo.Id, rsp.CoinCode, rsp.BatchNo, adds, t, tx)
		log.Infof("SaveBackAddress InsertBatchMchAddress err :%+v,\n", err)
		if err != nil {
			tx.Rollback()
			return
		}
	}
	if err != nil {
		tx.Rollback()
	} else {
		err = tx.Commit()
	}
	syncToAddrMgr(rsp.CoinCode, syncArr)
	return err
}

func syncToAddrMgr(coinCode string, addresses []string) {
	log.Infof("custody (新)拉取地址同步到addrmanagement codeCode=%s 地址=%v", coinCode, addresses)
	models := make([]entity.Addresses, 0)
	for _, a := range addresses {
		models = append(models, entity.Addresses{
			CreatedAt: time.Now(),
			Address:   a,
			CoinType:  coinCode,
			Status:    "used",
			ComeFrom:  "merchant",
			UserId:    5, // 对应 user表id 历史遗留问题，直接使用5
		})
	}
	rows, err := dao.AmAddBatchAddresses(models)
	if err != nil {
		log.Infof("custody 拉取地址同步到addrmanagement 插入到数据失败 %v", err)
		return
	}
	log.Infof("custody 拉取地址同步到addrmanagement 插入到数据受影响行数 %d", rows)
}
