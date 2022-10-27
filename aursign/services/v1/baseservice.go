package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eth-sign/conf"
	"github.com/eth-sign/model"
	"github.com/eth-sign/services"
	"github.com/eth-sign/util"
	log "github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"strings"
)

type generateKeyAndAddress func() (util.AddrInfo, error) //用于生成地址方法

type BaseService struct {
	*services.Service
}

func newBaseService() *BaseService {
	bs := new(BaseService)
	bs.Service = services.New()
	return bs
}

/*
传入地址或者公钥获取私钥
*/
func (bs *BaseService) addressOrPublicKeyToPrivate(publicKey string) (string, error) {
	return bs.GetKeyByAddress(publicKey)
}

func (bs *BaseService) createAddress(req *model.ReqCreateAddressParamsV2, generateKey generateKeyAndAddress) (*model.RespCreateAddressParams, error) {

	var addrInfos []util.AddrInfo
	for i := 0; i < req.Num; i++ {
		addrInfo, err := generateKey()
		if err != nil {
			return nil, err
		}
		addrInfos = append(addrInfos, addrInfo)
	}
	addresses, err := util.CreateAddrCsv(conf.Config.FilePath, req.MchId, req.OrderId, req.CoinName, addrInfos)
	if err != nil {
		return nil, err
	}
	resp := new(model.RespCreateAddressParams)
	resp.Address = addresses
	resp.BatchNo = req.OrderId
	resp.Mch = req.MchId
	resp.CoinCode = req.CoinName
	return resp, nil
}

func (bs *BaseService) multiThreadCreateAddress(number int, coinName, mchId, orderId string, generateKey generateKeyAndAddress) (*model.RespCreateAddressParams, error) {
	numcpu := runtime.NumCPU()
	buildnummap := []int{}
	addressChan := make(chan util.AddrInfo, number)
	if number <= numcpu {
		numcpu = 1
		buildnummap = append(buildnummap, number)
	} else {
		// 计算每个chan生成多少个
		avg := number / numcpu
		for j := 0; j < numcpu; j++ {
			buildnummap = append(buildnummap, avg)
		}
		buildnummap[numcpu-1] += (number % numcpu)
	}

	for i := 0; i < numcpu; i++ {
		buildnum := buildnummap[i]
		go func(workId int) {
			for index := 0; index < buildnum; index++ {
				addressInfo, err := generateKey()
				if err != nil {
					addressChan <- util.AddrInfo{
						Mnemonic: "",
						Address:  "",
						PrivKey:  "",
					}
					log.Errorf("work_id=[%d] create address error,Err=[%v]", workId, err)
					continue
				}
				addressChan <- addressInfo
			}
		}(i)
	}
	var addrInfos []util.AddrInfo
	total := 0
	for {
		select {
		case addrchan := <-addressChan:
			{
				total++
				if addrchan.Address == "" || addrchan.PrivKey == "" {
					break
				}
				addrInfos = append(addrInfos, addrchan)
			}
		}
		if total >= number {
			break
		}
	}
	if addrInfos == nil {
		return nil, errors.New("don`t have any address info")
	}
	log.Printf("Start write address to file,Create address number=[%d],Need address=[%d]", len(addrInfos), number)
	addresses, err := util.CreateAddrCsv(conf.Config.FilePath, mchId, orderId, conf.Config.CoinType, addrInfos)
	if err != nil {
		return nil, err
	}
	resp := new(model.RespCreateAddressParams)
	resp.Address = addresses
	resp.BatchNo = orderId
	resp.BatchNo = mchId
	resp.CoinCode = coinName
	return resp, nil
}

func (bs *BaseService) parseData(data, resp interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	log.Infof("请求参数为：%s", string(d))
	if err := json.Unmarshal(d, resp); err != nil {
		return err
	}
	return nil
}

func GetIService() services.IService {
	bs := newBaseService()
	prefix := strings.ToUpper(conf.Config.CoinType)
	name := fmt.Sprintf("%sService", prefix)
	return reflect.ValueOf(bs).MethodByName(name).Call(nil)[0].Interface().(services.IService)
}
