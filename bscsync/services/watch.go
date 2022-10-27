package services

import (
	"dataserver/log"
	"encoding/json"

	//"dataserver/log"
	"dataserver/models/bo"
	"dataserver/models/po"
	"fmt"
	"github.com/walletam/rabbitmq"
	"strings"
)

type WatchControl struct {
	CoinName                    string
	watchAddrs                  map[string][]bo.UserAddressInfo // *sync.Map
	watchUsers                  map[int64]po.UserInfo           // *sync.Map
	watchContracts              map[string]po.ContractInfo      // *sync.Map
	watchAddrNums, watchCttNums int
}

func NewWatchControl(name string) (*WatchControl, error) {
	s := &WatchControl{
		CoinName:       name,
		watchAddrs:     make(map[string][]bo.UserAddressInfo),
		watchUsers:     make(map[int64]po.UserInfo),
		watchContracts: make(map[string]po.ContractInfo),
	}
	return s, s.Reload()
}

func (s *WatchControl) Reload() error {
	log.Info("Watch Reload")
	userInfos, err := po.FindUserInfos()
	log.Infof("userInfos len:%d", len(userInfos))
	if err != nil {
		return fmt.Errorf("don't Find UserInfos, err :%v", err)
	}

	log.Info("Watch Reload 重新加载前数据=================")
	log.Infof("watchAddrs 数量 %d", len(s.watchAddrs))
	log.Infof("watchUsers 数量 %d", len(s.watchUsers))
	log.Infof("watchContracts 数量 %d", len(s.watchContracts))

	for _, userInfo := range userInfos {
		s.watchUsers[userInfo.ID] = userInfo
	}

	addrInfos, err := po.FindAddressesInfos(s.CoinName)
	log.Infof("addrInfos len:%d", len(addrInfos))

	if err != nil {
		return fmt.Errorf("don't Find AddressesInfos, err : %v", err)
	}
	for _, addrInfo := range addrInfos {
		userAddrInfo := bo.UserAddressInfo{
			UserID:  addrInfo.UserID,
			Address: addrInfo.Address,
		}
		// log.Infof("add addrInfos : %v", addrInfo)
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
	s.watchAddrNums = len(addrInfos)
	log.Infof("catch address len:[%d]", len(addrInfos))

	// todo 如果有智能合约
	if contractInfos, err := po.FindContractInfos(s.CoinName); err != nil {
		log.Errorf("FindContractInfos %v", err)
		return fmt.Errorf("don't Find ContractInfos, err :%v", err)
	} else {
		log.Infof("contractInfos len:%d", len(contractInfos))

		for _, contractInfo := range contractInfos {
			log.Debugf("add contractInfos : %v", contractInfo)
			s.watchContracts[strings.ToLower(contractInfo.ContractAddress)] = contractInfo
		}
	}
	log.Info("Watch Reload 重新加载后数据=================")
	log.Infof("watchAddrs 数量 %d", len(s.watchAddrs))
	log.Infof("watchUsers 数量 %d", len(s.watchUsers))
	log.Infof("watchContracts 数量 %d", len(s.watchContracts))
	return nil
}

func (s *WatchControl) ReloadWatchAddress() {
	log.Info("Start Reload Watch Address")
	num, err := po.FindAddressesNum(s.CoinName)
	if err != nil {
		log.Info("Reload Watch Address Error Code1: ", err.Error())
		return
	}
	if num == s.watchAddrNums {
		log.Infof("No new address added, no need to Reload. num: %d", num)
		return
	}
	userInfos, err := po.FindUserInfos()
	if err != nil {
		log.Info("Reload Watch Address Error Code2: ", err.Error())
		return
	}
	for _, userInfo := range userInfos {
		s.watchUsers[userInfo.ID] = userInfo
	}
	addrInfos, err := po.FindAddressesInfos(s.CoinName)
	if err != nil {
		log.Info("Reload Watch Address Error Code3: ", err.Error())
		return
	}
	//if len(addrInfos) == s.watchAddrNums {
	//	log.Infof("No new address added, no need to Reload. num: %d", len(addrInfos))
	//	return
	//}
	loadNum := 0
	for _, addrInfo := range addrInfos {
		if s.IsWatchAddressExist(addrInfo.Address) {
			continue
		}
		userAddrInfo := bo.UserAddressInfo{
			UserID:  addrInfo.UserID,
			Address: addrInfo.Address,
		}

		if v, ok := s.watchUsers[addrInfo.UserID]; ok {
			userAddrInfo.NotifyUrl = v.NotifyUrl
		}

		if v, ok := s.watchAddrs[strings.ToLower(addrInfo.Address)]; !ok {
			s.watchAddrs[strings.ToLower(addrInfo.Address)] = []bo.UserAddressInfo{userAddrInfo}
		} else {
			v = append(v, userAddrInfo)
			s.watchAddrs[strings.ToLower(addrInfo.Address)] = v
		}
		loadNum++
	}
	s.watchAddrNums = len(addrInfos)
	log.Infof("Reload watch address success, add num is %d, 当前地址数量: %d", loadNum, s.watchAddrNums)
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
		fmt.Println(err)
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
		fmt.Println(err)
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
	// log.Infof("wat address:%s, ressult:%t", addr, ok)
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
