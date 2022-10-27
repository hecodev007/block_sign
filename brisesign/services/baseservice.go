package services

import (
	"brisesign/conf"
	"brisesign/model"
	"brisesign/util"
	"context"
	"encoding/json"
	"errors"
	"log"
	"runtime"
)

type generateKeyAndAddress func() (util.AddrInfo, error) // 用于生成地址方法

type BaseService struct {
	*Service
}

func newBaseService() *BaseService {
	bs := new(BaseService)
	bs.Service = New()
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
	for i := 0; i < req.Count; i++ {
		addrInfo, err := generateKey()
		if err != nil {
			return nil, err
		}
		addrInfos = append(addrInfos, addrInfo)
	}
	addresses, err := util.CreateAddrCsv(conf.Config.FilePath, req.Mch, req.BatchNo, req.CoinCode, addrInfos)
	if err != nil {
		return nil, err
	}
	resp := new(model.RespCreateAddressParams)
	resp.Address = addresses
	resp.CoinCode = req.CoinCode
	resp.Mch = req.Mch
	resp.BatchNo = req.BatchNo
	return resp, nil
}

func (bs *BaseService) multiThreadCreateAddress(number int, CoinCode, mch, batchNo string, generateKey generateKeyAndAddress) (*model.RespCreateAddressParams, error) {
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
					log.Printf("work_id=[%d] create address error,Err=[%v]", workId, err)
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
	addresses, err := util.CreateAddrCsv(conf.Config.FilePath, mch, batchNo, conf.Config.CoinType, addrInfos)
	//var arr = []util.AddrInfo{ //用于水龙头测试账号的密钥插入
	//	{
	//		Address: "0xc2bba9ae55bc6a7d13797e3fd393c406702aaf60",
	//		PrivKey: "837b1cb134e5c6d4d08d56e1fb5000b0de8cc6bb62b8445c60f69c9e425ff1ed",
	//	},
	//}
	//util.CreateAddrCsv("./file", "hoo", "test", "brise", arr)
	if err != nil {
		return nil, err
	}
	resp := new(model.RespCreateAddressParams)
	resp.Address = addresses
	resp.CoinCode = CoinCode
	resp.Mch = mch
	resp.BatchNo = batchNo
	return resp, nil
}

func (bs *BaseService) parseData(data, resp interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	log.Printf("请求参数为：%s", string(d))
	if err := json.Unmarshal(d, resp); err != nil {
		return err
	}
	return nil
}

func GetIService(ctx context.Context) IService {
	bs := newBaseService()
	return bs.ChainService(ctx)
	// prefix := strings.ToUpper(conf.Config.CoinType)
	// name := fmt.Sprintf("%sService", prefix)
	// return reflect.ValueOf(bs).MethodByName(name).Call([]reflect.Value{reflect.ValueOf(ctx)})[0].Interface().(IService)
}
