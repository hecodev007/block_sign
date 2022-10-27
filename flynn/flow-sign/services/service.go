package services

import (
	"encoding/csv"
	"fmt"
	"github.com/group-coldwallet/flow-sign/common"
	"github.com/group-coldwallet/flow-sign/conf"
	"github.com/group-coldwallet/flow-sign/model"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type IService interface {
	SignService(req *model.ReqSignParams) (interface{}, error)
	CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error)
	TransferService(req interface{}) (interface{}, error)
	MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error
	GetBalance(req *model.ReqGetBalanceParams) (interface{}, error)
	ValidAddress(address string) error
}

type Service struct {
	aesKeyMap     map[string]string
	encryptKeyMap map[string]string
}

func New() *Service {
	s := new(Service)
	s.aesKeyMap = make(map[string]string)
	s.encryptKeyMap = make(map[string]string)
	s.initKeyMap()
	return s
}

func (s *Service) initKeyMap() {
	paths := s.loadKeyFilePath()
	if paths == nil || len(paths) == 0 {
		//log.Error("Load private key error,maybe is can find .csv,please check you path")
		return
	}
	err := s.loadKeyFile(paths)
	if err != nil {

	}

}

func (s *Service) loadKeyFile(paths []string) error {

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		// 这个方法体执行完成后，关闭文件
		defer file.Close()
		reader := csv.NewReader(file)
		//reader.FieldsPerRecord = -1
		//idx:=1
		for {
			//log.Println(path,idx)
			// Read返回的是一个数组，它已经帮我们分割了，
			record, err := reader.Read()
			// 如果读到文件的结尾，EOF的优先级比nil高！
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println(path)
				log.Println(reader.FieldsPerRecord)
				log.Error("记录集错误:", err)
				return nil
			}
			if strings.Contains(path, "_a_") {
				s.encryptKeyMap[record[0]] = record[1]
			} else {
				s.aesKeyMap[record[0]] = record[1]
			}
			//idx++
		}
	}
	return nil
}
func (s *Service) loadKeyFilePath() []string {
	var (
		paths []string
	)

	filepath.Walk(conf.Config.FilePath, func(path string, info os.FileInfo, err error) error {

		if strings.Contains(path, ".csv") {
			if strings.Contains(path, "_a_") || strings.Contains(path, "_b_") {
				paths = append(paths, path)
			}
		}
		return nil
	})
	return paths
}

func (s *Service) GetAesKeyMap() map[string]string {
	return s.aesKeyMap
}

func (s *Service) GetEncryptKeyMap() map[string]string {
	return s.encryptKeyMap
}

func (s *Service) GetKeyByAddress(address string) (string, error) {
	aesKey := s.aesKeyMap[address]
	encryptKey := s.encryptKeyMap[address]
	if aesKey == "" || encryptKey == "" {
		return "", fmt.Errorf("Load aes key or encrypt key is null,AES=[%s],ENCRYPT=[%s],Address=[%s]", aesKey,
			encryptKey, address)
	}
	privateBytes, err := common.AesBase64Crypt([]byte(encryptKey), []byte(aesKey), false)
	if err != nil {
		return "", err
	}
	wif := string(privateBytes)
	return wif, nil
}
