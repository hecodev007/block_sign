package services

import (
	"encoding/json"
	"fmt"
	"log"
	"rsksync/models/bo"
	"rsksync/models/po"
	//"rsksync/common/log"
	"github.com/walletam/rabbitmq"
	"strings"
	"time"
)

type WatchControl struct {
	CoinName           string
	AddrTableMaxid     int64
	ContractTableMaxId int64
	watchAddrs         map[string][]bo.UserAddressInfo //*sync.Map
	watchUsers         map[int64]po.UserInfo           //*sync.Map
	watchContracts     map[string]po.ContractInfo      //*sync.Map
}

var WatchCtl *WatchControl

func NewWatchControl(name string, addressRecover, contractRecover int64) (*WatchControl, error) {
	s := &WatchControl{
		CoinName:       name,
		watchAddrs:     make(map[string][]bo.UserAddressInfo),
		watchUsers:     make(map[int64]po.UserInfo),
		watchContracts: make(map[string]po.ContractInfo),
	}

	userInfos, err := po.FindUserInfos()
	if err != nil {
		return nil, fmt.Errorf("don't Find UserInfos, err :%v", err)
	}
	for _, userInfo := range userInfos {
		s.watchUsers[userInfo.ID] = userInfo
	}
	s.AddrTableMaxid, err = new(po.AddressesInfo).MaxId()
	if err != nil {
		panic(err.Error())
	}
	s.ContractTableMaxId, err = new(po.ContractInfo).MaxId()
	if err != nil {
		panic(err.Error())
	}
	addrInfos, err := po.FindAddressesInfos(s.CoinName, 0)
	if err != nil {
		return nil, fmt.Errorf("don't Find AddressesInfos, err : %v", err)
	}
	for _, addrInfo := range addrInfos {
		userAddrInfo := bo.UserAddressInfo{
			UserID:  addrInfo.UserID,
			Address: addrInfo.Address,
		}
		//log.Infof("add addrInfos : %v", addrInfo)
		if v, ok := s.watchUsers[addrInfo.UserID]; ok {
			userAddrInfo.NotifyUrl = v.NotifyUrl
		}

		if v, ok := s.watchAddrs[strings.ToLower(addrInfo.Address)]; !ok {
			s.watchAddrs[strings.ToLower(addrInfo.Address)] = []bo.UserAddressInfo{userAddrInfo}
		} else {
			v = append(v, userAddrInfo)
			s.watchAddrs[strings.ToLower(addrInfo.Address)] = v
		}
	}

	log.Printf("catch address len:[%d]", len(addrInfos))

	//todo 如果有智能合约
	if contractInfos, err := po.FindContractInfos(s.CoinName, 0); err != nil {
		log.Printf("FindContractInfos %v", err)
		return nil, fmt.Errorf("don't Find ContractInfos, err :%v", err)
	} else {
		for k, contractInfo := range contractInfos {
			cName := strings.ToLower(contractInfo.ContractAddress)
			log.Printf("add contractInfos : %v", contractInfo)

			//表设计问题，硬编码实现
			if strings.HasPrefix(contractInfo.ContractAddress, "bsc:") {
				contractInfos[k].ContractAddress = strings.ReplaceAll(cName, "bsc:", "")
				s.watchContracts[cName] = contractInfos[k]
			} else {
				s.watchContracts[cName] = contractInfo
			}

		}
		//由于多链的出现合约不能再是唯一键，硬编码添加合约
	}

	if addressRecover > 0 {
		go s.addressDiscover(addressRecover)
	}
	if contractRecover > 0 {
		go s.contractDiscover(contractRecover)
	}
	return s, nil
}
func (s *WatchControl) addressDiscover(sleepSecond int64) {
	for {

		func() {
			st := time.Now()
			maxid, err := new(po.AddressesInfo).MaxId()
			if err != nil {
				log.Println(err.Error())
				return
			}
			if maxid == s.AddrTableMaxid {
				return
			}
			addrList, err := po.FindAddressesInfos(s.CoinName, s.AddrTableMaxid)
			if err != nil {
				log.Println(err.Error())
				return
			}
			s.AddrTableMaxid = maxid

			if len(addrList) == 0 {
				return
			}
			//log.Infof("新增地址查询! 耗时:%v", len(addrList), time.Since(st))
			watchAddrs := make(map[string][]bo.UserAddressInfo)
			for k, v := range s.watchAddrs {
				watchAddrs[k] = v
			}

			for _, addrInfo := range addrList {
				userAddrInfo := bo.UserAddressInfo{
					UserID:  addrInfo.UserID,
					Address: addrInfo.Address,
				}
				if v, ok := s.watchUsers[addrInfo.UserID]; ok {
					userAddrInfo.NotifyUrl = v.NotifyUrl
				}

				if v, ok := watchAddrs[strings.ToLower(addrInfo.Address)]; !ok {
					watchAddrs[strings.ToLower(addrInfo.Address)] = []bo.UserAddressInfo{userAddrInfo}
				} else {
					v = append(v, userAddrInfo)
					watchAddrs[strings.ToLower(addrInfo.Address)] = v
				}
			}
			s.watchAddrs = watchAddrs
			log.Printf("新增地址%v成功! 耗时:%v", len(addrList), time.Since(st))
			return
		}()
		time.Sleep(time.Second * time.Duration(sleepSecond))
	}

}
func (s *WatchControl) contractDiscover(sleepSecond int64) {
	for {
		func() {
			st := time.Now()
			maxid, err := new(po.ContractInfo).MaxId()
			if err != nil {
				log.Println(err.Error())
				return
			}
			if maxid == s.ContractTableMaxId {
				return
			}
			contractInfos, err := po.FindContractInfos(s.CoinName, s.ContractTableMaxId)
			if err != nil {
				log.Println(err.Error())
				return
			}
			s.ContractTableMaxId = maxid
			if len(contractInfos) == 0 {
				return
			}
			watchContracts := make(map[string]po.ContractInfo)
			for k, v := range s.watchContracts {
				watchContracts[k] = v
			}

			for _, contractInfo := range contractInfos {
				watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
			}

			s.watchContracts = watchContracts
			log.Printf("新增合约%v成功! 耗时:%v", len(contractInfos), time.Since(st))
		}()

		time.Sleep(time.Second * time.Duration(sleepSecond))
	}

}
func (s *WatchControl) InsertWatchAddress(userid int64, address, notifyurl string) {
	userAddrInfo := bo.UserAddressInfo{
		UserID:    userid,
		Address:   address,
		NotifyUrl: notifyurl,
	}

	if v, ok := s.watchAddrs[strings.ToLower(address)]; !ok {
		s.watchAddrs[strings.ToLower(address)] = []bo.UserAddressInfo{userAddrInfo}
	} else {
		v = append(v, userAddrInfo)
		s.watchAddrs[strings.ToLower(address)] = v
	}
}

func (s *WatchControl) InsertWatchContract(name, contractaddr, cointype string, decimal int) {
	contractInfo := po.ContractInfo{
		Name:            name,
		ContractAddress: contractaddr,
		Decimal:         decimal,
		CoinType:        cointype,
	}
	s.watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
}
func (s *WatchControl) InsertContract(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool {
	//如果返回true 则无需重试
	fmt.Printf("data:%s\n", string(data))

	contractInfos := make([]po.ContractInfo, 0)
	err := json.Unmarshal(data, &contractInfos)
	if err != nil {
		log.Println(err)
		return false
	}
	addrs := make(map[string]po.ContractInfo)
	for k, v := range s.watchContracts {
		addrs[k] = v
	}
	//name, contractaddr, cointype string, decimal int
	for _, contractInfo := range contractInfos {
		contractInfo := po.ContractInfo{
			Name:            contractInfo.Name,
			ContractAddress: contractInfo.ContractAddress,
			Decimal:         contractInfo.Decimal,
			CoinType:        contractInfo.CoinType,
		}
		addrs[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
		//s.InsertWatchContract(contractInfo.Name, contractInfo.ContractAddress, contractInfo.CoinType, contractInfo.Decimal)
	}
	s.watchContracts = addrs
	return true
}
func (s *WatchControl) InsertAddr(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool {
	//如果返回true 则无需重试
	fmt.Printf("data:%s\n", string(data))

	addrInfos := make([]bo.UserAddressInfo, 0)
	err := json.Unmarshal(data, &addrInfos)
	if err != nil {
		log.Println(err)
		return false
	}
	addrs := make(map[string][]bo.UserAddressInfo)
	for k, v := range s.watchAddrs {
		addrs[k] = v
	}
	for _, addrInfo := range addrInfos {
		userAddrInfo := bo.UserAddressInfo{
			UserID:    addrInfo.UserID,
			Address:   addrInfo.Address,
			NotifyUrl: addrInfo.NotifyUrl,
		}

		if v, ok := addrs[strings.ToLower(addrInfo.Address)]; !ok {
			addrs[strings.ToLower(addrInfo.Address)] = []bo.UserAddressInfo{userAddrInfo}
		} else {
			v = append(v, userAddrInfo)
			addrs[strings.ToLower(addrInfo.Address)] = v
		}
		s.watchAddrs = addrs
		//s.InsertWatchAddress(addrInfo.UserID, addrInfo.Address, addrInfo.NotifyUrl)
	}

	return true
}
func (s *WatchControl) RemoveWatchContract(contractaddr string) error {
	if _, ok := s.watchContracts[strings.ToLower(contractaddr)]; ok {
		delete(s.watchContracts, strings.ToLower(contractaddr))
	} else {
		return fmt.Errorf("don't find contract %s", contractaddr)
	}
	return nil
}

func (s *WatchControl) IsWatchAddressExist(addr string) bool {

	_, ok := s.watchAddrs[strings.ToLower(addr)]
	//log.Infof("wat address:%s, ressult:%t", addr, ok)
	return ok
}

func (s *WatchControl) IsWatchUserExist(userId int64) bool {
	_, ok := s.watchUsers[userId]
	return ok
}

func (s *WatchControl) IsContractExist(addr string) bool {
	_, ok := s.watchContracts[strings.ToLower(addr)]
	return ok
}

func (s *WatchControl) GetContract(addr string) (po.ContractInfo, error) {
	v, ok := s.watchContracts[strings.ToLower(addr)]
	if !ok {
		return po.ContractInfo{}, fmt.Errorf("don't find contract : %s", addr)
	}
	return v, nil
}

func (s *WatchControl) GetWatchAddress(addr string) ([]bo.UserAddressInfo, error) {
	v, ok := s.watchAddrs[strings.ToLower(addr)]
	if !ok {
		return nil, fmt.Errorf("don't find addresses info : %s", addr)
	}
	return v, nil
}

func (s *WatchControl) GetWatchUserNotifyUrl(userId int64) (string, error) {
	v, ok := s.watchUsers[userId]
	if !ok {
		return "", fmt.Errorf("don't find user info :%d", userId)
	}
	return v.NotifyUrl, nil
}

func (s *WatchControl) GetContractNameAndDecimal(addr string) (string, int, error) {
	contractInfo, ok := s.watchContracts[strings.ToLower(addr)]
	if !ok {
		return "", 0, fmt.Errorf("don't find contract : %s", addr)
	}
	return contractInfo.Name, contractInfo.Decimal, nil
}

func (s *WatchControl) AddWatchContract() {

}

func (s *WatchControl) UpdateWatchAddress(req *bo.UpdateRequest) {

}

func (s *WatchControl) RemoveWatchAddress(req *bo.RemoveRequest) {

}
