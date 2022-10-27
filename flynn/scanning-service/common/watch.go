package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/scanning-service/log"
	"github.com/group-coldwallet/scanning-service/models/bo"
	"github.com/group-coldwallet/scanning-service/models/po"
	"github.com/walletam/rabbitmq"
	"strings"
	"sync"
)

type WatchControl struct {
	CoinName                    string
	watchAddresses              *sync.Map
	watchUser                   *sync.Map
	watchContracts              *sync.Map
	watchAddrNums, watchCttNums int
}

func (s *WatchControl) getWatchUserValue(key interface{}) (*po.UserInfo, error) {
	if s.watchUser == nil {
		return nil, errors.New("watch user is nil ptr")
	}
	ui, isExist := s.watchUser.Load(key)
	if !isExist {
		return nil, fmt.Errorf("do not exist this user info,key=%v", key)
	}
	if ui == nil {
		return nil, fmt.Errorf("get user info is null,key=%v", key)
	}
	d, _ := json.Marshal(ui)
	var userInfo po.UserInfo
	err := json.Unmarshal(d, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal user info error,err=%v", err)
	}
	return &userInfo, nil
}

func (s *WatchControl) getContractValue(key interface{}) (*po.ContractInfo, error) {
	if s.watchContracts == nil {
		return nil, errors.New("contract info is nil ptr")
	}
	ui, isExist := s.watchContracts.Load(key)
	if !isExist {
		return nil, fmt.Errorf("do not exist this contract info,key=%v", key)
	}
	if ui == nil {
		return nil, fmt.Errorf("get contract info is null,key=%v", key)
	}
	d, _ := json.Marshal(ui)
	var contractInfo po.ContractInfo
	err := json.Unmarshal(d, &contractInfo)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal contract info error,err=%v", err)
	}
	return &contractInfo, nil
}

func (s *WatchControl) getWatchAddressValue(key interface{}) ([]bo.UserAddressInfo, error) {
	if s.watchAddresses == nil {
		return nil, errors.New("watch address is nil ptr")
	}
	wa, isExist := s.watchAddresses.Load(key)
	if !isExist {
		return nil, fmt.Errorf("do not exist this address info,key=%v", key)
	}
	if wa == nil {
		return nil, fmt.Errorf("get address info is null,key=%v", key)
	}
	d, _ := json.Marshal(wa)
	var addrInfos []bo.UserAddressInfo
	err := json.Unmarshal(d, &addrInfos)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal address info error,err=%v", err)
	}
	return addrInfos, nil
}

func NewWatchControl(name string) (*WatchControl, error) {
	s := &WatchControl{
		CoinName:       name,
		watchAddresses: &sync.Map{},
		watchUser:      &sync.Map{},
		watchContracts: &sync.Map{},
	}
	userInfos, err := po.FindUserInfos()
	if err != nil {
		return nil, fmt.Errorf("don't find user infos, err :%v", err)
	}
	for _, userInfo := range userInfos {
		s.watchUser.Store(userInfo.ID, userInfo)
	}

	addrInfos, err := po.FindAddressesInfos(s.CoinName)
	if err != nil {
		return nil, fmt.Errorf("don't find address infos, err : %v", err)
	}
	for _, addrInfo := range addrInfos {
		userAddrInfo := bo.UserAddressInfo{
			UserID:  addrInfo.UserID,
			Address: addrInfo.Address,
		}
		ui, err := s.getWatchUserValue(addrInfo.UserID)
		if err == nil {
			userAddrInfo.NotifyUrl = ui.NotifyUrl
		}
		wa, err := s.getWatchAddressValue(strings.ToLower(addrInfo.Address))
		if err != nil {
			s.watchAddresses.Store(strings.ToLower(addrInfo.Address), []bo.UserAddressInfo{userAddrInfo})
		} else {
			wa = append(wa, userAddrInfo)
			s.watchAddresses.Store(strings.ToLower(addrInfo.Address), wa)
		}
	}
	log.Infof("watch address success,watch num is %d", len(addrInfos))
	s.watchAddrNums = len(addrInfos)
	contractInfos, err := po.FindContractInfos(s.CoinName)
	if err != nil {
		log.Errorf("find contract info error,Err=%v", err)
		return nil, fmt.Errorf("don't find contract info ,err: %v", err)
	} else {
		for _, contractInfo := range contractInfos {
			log.Infof("add contractInfos : %v", contractInfo)
			s.watchContracts.Store(strings.ToLower(contractInfo.ContractAddress), contractInfo)
		}
	}
	log.Infof("watch contract success,watch num is %d", len(contractInfos))
	s.watchCttNums = len(contractInfos)
	return s, nil
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
		s.watchUser.Store(userInfo.ID, userInfo)
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
		ui, err := s.getWatchUserValue(addrInfo.UserID)
		if err == nil {
			userAddrInfo.NotifyUrl = ui.NotifyUrl
		}
		wa, err := s.getWatchAddressValue(strings.ToLower(addrInfo.Address))
		if err != nil {
			s.watchAddresses.Store(strings.ToLower(addrInfo.Address), []bo.UserAddressInfo{userAddrInfo})
		} else {
			wa = append(wa, userAddrInfo)
			s.watchAddresses.Store(strings.ToLower(addrInfo.Address), wa)
		}
		loadNum++
	}
	s.watchAddrNums = len(addrInfos)
	log.Infof("Reload watch address success, add num is %d, 当前地址数量: %d", loadNum, s.watchAddrNums)
}

func (s *WatchControl) InsertWatchAddress(userid int64, address, notifyurl string) {
	// 判断地址是否存在内存中
	if s.IsWatchAddressExist(address) {
		log.Infof("该地址[%s]已经存在于内存中", address)
		return
	}
	userAddrInfo := bo.UserAddressInfo{
		UserID:    userid,
		Address:   address,
		NotifyUrl: notifyurl,
	}
	wa, err := s.getWatchAddressValue(strings.ToLower(address))
	if err != nil {
		s.watchAddresses.Store(strings.ToLower(address), []bo.UserAddressInfo{userAddrInfo})
	} else {
		wa = append(wa, userAddrInfo)
		s.watchAddresses.Store(strings.ToLower(address), wa)
	}
	s.watchAddrNums++
}

func (s *WatchControl) InsertWatchContract(name, contractaddr, cointype string, decimal int) {
	if s.IsContractExist(contractaddr) {
		log.Infof("该合约地址[%s]已经存在于内存中", contractaddr)
		return
	}
	contractInfo := po.ContractInfo{
		Name:            name,
		ContractAddress: contractaddr,
		Decimal:         decimal,
		CoinType:        cointype,
	}
	s.watchContracts.Store(strings.ToLower(contractInfo.ContractAddress), contractInfo)
	s.watchCttNums++
}

func (s *WatchControl) InsertContract(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool {
	//如果返回true 则无需重试
	fmt.Printf("data:%s\n", string(data))

	contractInfos := make([]po.ContractInfo, 0)
	err := json.Unmarshal(data, &contractInfos)
	if err != nil {
		log.Error(err)
		return false
	}
	//addrs := make(map[string]*po.ContractInfo)
	//for k, v := range s.watchContracts {
	//	addrs[k] = v
	//}
	//name, contractaddr, cointype string, decimal int
	for _, contractInfo := range contractInfos {

		s.InsertWatchContract(contractInfo.Name, contractInfo.ContractAddress, contractInfo.CoinType, contractInfo.Decimal)
	}

	return true
}
func (s *WatchControl) InsertAddr(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool {
	//如果返回true 则无需重试
	fmt.Printf("data:%s\n", string(data))

	addrInfos := make([]bo.UserAddressInfo, 0)
	err := json.Unmarshal(data, &addrInfos)
	if err != nil {
		log.Error(err)
		return false
	}
	//addrs := make(map[string][]*bo.UserAddressInfo)
	//for k, v := range s.watchContracts {
	//	addrs[k] = v
	//}
	for _, addrInfo := range addrInfos {

		s.InsertWatchAddress(addrInfo.UserID, addrInfo.Address, addrInfo.NotifyUrl)
	}

	return true
}
func (s *WatchControl) RemoveWatchContract(contractaddr string) error {
	if !s.IsContractExist(strings.ToLower(contractaddr)) {
		return nil
	}
	s.watchContracts.Delete(strings.ToLower(contractaddr))
	s.watchCttNums--
	return nil
}

func (s *WatchControl) IsWatchAddressExist(addr string) bool {
	_, err := s.getWatchAddressValue(strings.ToLower(addr))
	if err != nil {
		//log.Infof("IsWatchAddressNotExist addr: %s, err: %s",addr,  err.Error())
		return false
	}
	return true
}

func (s *WatchControl) IsWatchUserExist(userId int64) bool {
	_, err := s.getWatchUserValue(userId)
	if err != nil {
		log.Errorf("get watch user info error,err: %v", err)
		return false
	}
	return true
}

func (s *WatchControl) IsContractExist(addr string) bool {
	_, err := s.getContractValue(strings.ToLower(addr))
	if err != nil {
		//log.Errorf("get watch contract info error,err: %v",err)
		return false
	}
	return true
}
func (s *WatchControl) GetContract(addr string) (*po.ContractInfo, error) {
	wc, err := s.getContractValue(strings.ToLower(addr))
	if err != nil {
		return nil, fmt.Errorf("get watch contract info error,err: %v", err)
	}
	return wc, nil
}
func (s *WatchControl) GetWatchAddress(addr string) ([]bo.UserAddressInfo, error) {
	wa, err := s.getWatchAddressValue(strings.ToLower(addr))
	if err != nil {
		return nil, fmt.Errorf("get watch address error err: %v", err)
	}
	return wa, nil
}

func (s *WatchControl) GetWatchUserNotifyUrl(userId int64) (string, error) {
	wu, err := s.getWatchUserValue(userId)
	if err != nil {
		return "", fmt.Errorf("get watch user info error,err: %v", err)
	}
	return wu.NotifyUrl, nil
}

func (s *WatchControl) RemoveWatchAddress(req *bo.RemoveRequest) {
	if !s.IsWatchAddressExist(req.Address) {
		return
	}
	ais, err := s.getWatchAddressValue(strings.ToLower(req.Address))
	if err != nil {
		return
	}
	var newAis []bo.UserAddressInfo
	for _, ai := range ais {
		if ai.UserID == req.UserId {
			continue
		}
		newAis = append(newAis, ai)
	}
	if len(newAis) > 0 {
		s.watchAddresses.Store(strings.ToLower(req.Address), newAis)
	} else {
		s.watchAddresses.Delete(strings.ToLower(req.Address))
		s.watchAddrNums--
	}
}

func (s *WatchControl) UpdateWatchAddress(req *bo.UpdateRequest) {
	//if s.IsWatchUserExist(req.UserId) {
	//	ui,err:=s.getWatchUserValue(req.UserId)
	//	if err != nil {
	//		return
	//	}
	//	ui.NotifyUrl = req.Url
	//	s.watchUser.Store(req.UserId,ui)
	//}
	//
	//s.watchAddresses.Range(func(key, value interface{}) bool {
	//	if value!=nil {
	//		d, _ := json.Marshal(value)
	//		var addrInfos []bo.UserAddressInfo
	//		err := json.Unmarshal(d, &addrInfos)
	//		if err == nil {
	//			for _,ai:=range addrInfos{
	//				if ai.UserID==req.UserId {
	//
	//				}
	//			}
	//		}
	//	}
	//})
	// 暂不支持
	return
}

func (s *WatchControl) GetWatchAddressNums() int {
	return s.watchAddrNums
}

func (s *WatchControl) GetWatchContractNums() int {
	return s.watchCttNums
}
