package services

import (
	"fmt"
	"oktsync/common/log"
	"oktsync/models/bo"
	"oktsync/models/po"
	"strings"
	"time"
)

type WatchControl struct {
	CoinName           string
	AddrTableMaxid     int64
	ContractTableMaxId int64
	WatchAddrs         map[string][]*bo.UserAddressInfo //*sync.Map
	watchUsers         map[int64]*po.UserInfo           //*sync.Map
	WatchContracts     map[string]*po.ContractInfo      //*sync.Map
}

func NewWatchControl(name string, addressRecover, contractRecover int64) *WatchControl {

	s := &WatchControl{
		CoinName:       name,
		WatchAddrs:     make(map[string][]*bo.UserAddressInfo),
		watchUsers:     make(map[int64]*po.UserInfo),
		WatchContracts: make(map[string]*po.ContractInfo),
	}

	userInfos, err := po.FindUserInfos()
	if err != nil {
		panic(err.Error())
	}
	for _, userInfo := range userInfos {
		s.watchUsers[userInfo.Id] = userInfo
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
		panic(err.Error())
	}
	for _, addrInfo := range addrInfos {
		userAddrInfo := bo.UserAddressInfo{
			UserID:  addrInfo.UserId,
			Address: addrInfo.Address,
		}
		if v, ok := s.watchUsers[addrInfo.UserId]; ok {
			userAddrInfo.NotifyUrl = v.TrxNotifyUrl
		}

		if v, ok := s.WatchAddrs[strings.ToLower(addrInfo.Address)]; !ok {
			s.WatchAddrs[strings.ToLower(addrInfo.Address)] = []*bo.UserAddressInfo{&userAddrInfo}
		} else {
			v = append(v, &userAddrInfo)
			s.WatchAddrs[strings.ToLower(addrInfo.Address)] = v
		}
	}
	//addrsjson, _ := json.Marshal(addrInfos)
	log.Infof("监控地址数:[%v]", len(addrInfos))

	//todo 如果有智能合约
	if contractInfos, err := po.FindContractInfos(s.CoinName, 0); err != nil {
		panic(err.Error())
	} else {
		log.Infof("监控合约数:[%v]", len(contractInfos))
		for _, contractInfo := range contractInfos {
			//log.Info("add contractInfos : ", contractInfo.Name,contractInfo.ContractAddress)
			s.WatchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
		}
	}

	if addressRecover > 0 {
		go s.addressDiscover(addressRecover)
	}
	if contractRecover > 0 {
		go s.contractDiscover(contractRecover)
	}
	return s
}
func (s *WatchControl) addressDiscover(sleepSecond int64) {
	for {

		func() {
			st := time.Now()
			maxid, err := new(po.AddressesInfo).MaxId()
			if err != nil {
				log.Error(err.Error())
				return
			}
			if maxid == s.AddrTableMaxid {
				return
			}
			addrList, err := po.FindAddressesInfos(s.CoinName, s.AddrTableMaxid)
			if err != nil {
				log.Info(err.Error())
				return
			}
			s.AddrTableMaxid = maxid

			if len(addrList) == 0 {
				return
			}
			//log.Infof("新增地址查询! 耗时:%v", len(addrList), time.Since(st))
			watchAddrs := make(map[string][]*bo.UserAddressInfo)
			for k, v := range s.WatchAddrs {
				watchAddrs[k] = v
			}

			for _, addrInfo := range addrList {
				userAddrInfo := bo.UserAddressInfo{
					UserID:  addrInfo.UserId,
					Address: addrInfo.Address,
				}
				if v, ok := s.watchUsers[addrInfo.UserId]; ok {
					userAddrInfo.NotifyUrl = v.TrxNotifyUrl
				}

				if v, ok := watchAddrs[strings.ToLower(addrInfo.Address)]; !ok {
					watchAddrs[strings.ToLower(addrInfo.Address)] = []*bo.UserAddressInfo{&userAddrInfo}
				} else {
					v = append(v, &userAddrInfo)
					watchAddrs[strings.ToLower(addrInfo.Address)] = v
				}
			}
			s.WatchAddrs = watchAddrs
			log.Infof("新增地址%v成功! 耗时:%v", len(addrList), time.Since(st))
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
				log.Error(err.Error())
				return
			}
			if maxid == s.ContractTableMaxId {
				return
			}
			contractInfos, err := po.FindContractInfos(s.CoinName, s.ContractTableMaxId)
			if err != nil {
				log.Error(err.Error())
				return
			}
			s.ContractTableMaxId = maxid
			if len(contractInfos) == 0 {
				return
			}
			watchContracts := make(map[string]*po.ContractInfo)
			for k, v := range s.WatchContracts {
				watchContracts[k] = v
			}

			for _, contractInfo := range contractInfos {
				watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
			}

			s.WatchContracts = watchContracts
			log.Infof("新增合约%v成功! 耗时:%v", len(contractInfos), time.Since(st))
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

	if v, ok := s.WatchAddrs[strings.ToLower(address)]; !ok {
		s.WatchAddrs[strings.ToLower(address)] = []*bo.UserAddressInfo{&userAddrInfo}
	} else {
		v = append(v, &userAddrInfo)
		s.WatchAddrs[strings.ToLower(address)] = v
	}
}

func (s *WatchControl) InsertWatchContract(name, contractaddr, cointype string, decimal int) {
	contractInfo := &po.ContractInfo{
		Name:            name,
		ContractAddress: contractaddr,
		Decimal:         decimal,
		CoinType:        cointype,
	}
	s.WatchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
}

func (s *WatchControl) RemoveWatchContract(contractaddr string) error {
	if _, ok := s.WatchContracts[strings.ToLower(contractaddr)]; ok {
		delete(s.WatchContracts, strings.ToLower(contractaddr))
	} else {
		return fmt.Errorf("don't find contract %s", contractaddr)
	}
	return nil
}

func (s *WatchControl) IsWatchAddressExist(addr string) bool {

	_, ok := s.WatchAddrs[strings.ToLower(addr)]
	if ok {
		//log.Infof("wat address:%s, ressult:%t", addr, ok)
	}
	return ok
}

func (s *WatchControl) IsWatchUserExist(userId int64) bool {
	_, ok := s.watchUsers[userId]
	return ok
}

func (s *WatchControl) IsContractExist(addr string) bool {
	_, ok := s.WatchContracts[strings.ToLower(addr)]
	if ok {
		//log.Infof("IsContractExist:%s, ressult:%t", addr, ok)
	}
	return ok
}

func (s *WatchControl) GetContract(addr string) (*po.ContractInfo, error) {
	v, ok := s.WatchContracts[strings.ToLower(addr)]
	if !ok {
		return new(po.ContractInfo), fmt.Errorf("don't find contract : %s", addr)
	}
	return v, nil
}

func (s *WatchControl) GetWatchAddress(addr string) ([]*bo.UserAddressInfo, error) {
	v, ok := s.WatchAddrs[strings.ToLower(addr)]
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
	return v.TrxNotifyUrl, nil
}

func (s *WatchControl) GetContractNameAndDecimal(addr string) (string, int, error) {
	contractInfo, ok := s.WatchContracts[strings.ToLower(addr)]
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
