package services

import (
	"biwDataServer/common/log"
	"biwDataServer/models/bo"
	"biwDataServer/models/po"
	"encoding/json"
	"fmt"
	"strings"
)

type WatchControl struct {
	CoinName       string
	WatchAddrs     map[string][]*bo.UserAddressInfo //*sync.Map
	watchUsers     map[int64]*po.UserInfo           //*sync.Map
	watchContracts map[string]*po.ContractInfo      //*sync.Map
}

func NewWatchControl(name string) *WatchControl {
	s := &WatchControl{
		CoinName:       name,
		WatchAddrs:     make(map[string][]*bo.UserAddressInfo),
		watchUsers:     make(map[int64]*po.UserInfo),
		watchContracts: make(map[string]*po.ContractInfo),
	}

	userInfos, err := po.FindUserInfos()
	if err != nil {
		panic(err.Error())
	}
	for _, userInfo := range userInfos {
		s.watchUsers[userInfo.Id] = userInfo
	}

	addrInfos, err := po.FindAddressesInfos(s.CoinName)
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
	addrsjson, _ := json.Marshal(addrInfos)
	log.Infof("catch address number:[%d] %v", len(addrInfos), string(addrsjson))

	//todo 如果有智能合约
	if contractInfos, err := po.FindContractInfos(s.CoinName); err != nil {
		panic(err.Error())
	} else {
		for _, contractInfo := range contractInfos {
			log.Infof("add contractInfos : %v", contractInfo)
			s.watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
		}
	}

	return s
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

	_, ok := s.WatchAddrs[strings.ToLower(addr)]
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
	if ok {
		log.Infof("IsContractExist:%s, ressult:%t", addr, ok)
	}
	return ok
}

func (s *WatchControl) GetContract(addr string) (*po.ContractInfo, error) {
	v, ok := s.watchContracts[strings.ToLower(addr)]
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
