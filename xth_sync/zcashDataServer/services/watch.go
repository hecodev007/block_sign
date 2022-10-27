package services

import (
	"fmt"
	"strings"
	"zcashDataServer/common/log"
	"zcashDataServer/models/bo"
	"zcashDataServer/models/po"
)

type WatchControl struct {
	CoinName       string
	watchAddrs     map[string][]bo.UserAddressInfo //*sync.Map
	watchUsers     map[int64]po.UserInfo           //*sync.Map
	watchContracts map[string]po.ContractInfo      //*sync.Map
}

func NewWatchControl(name string) (*WatchControl, error) {
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

	addrInfos, err := po.FindAddressesInfos(s.CoinName)
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

	log.Infof("catch address len:[%d]", len(addrInfos))

	//todo 如果有智能合约
	if contractInfos, err := po.FindContractInfos(s.CoinName); err != nil {
		log.Errorf("FindContractInfos %v", err)
		return nil, fmt.Errorf("don't Find ContractInfos, err :%v", err)
	} else {
		for _, contractInfo := range contractInfos {
			log.Infof("add contractInfos : %v", contractInfo)
			s.watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
		}
	}

	return s, nil
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
	if ok {
		log.Infof("wat address:%s, ressult:%t", addr, ok)
	}
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
