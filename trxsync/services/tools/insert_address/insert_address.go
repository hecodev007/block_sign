package insert_address

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/group-coldwallet/trxsync/conf"
	"github.com/group-coldwallet/trxsync/models/po"
	"io"
	"log"
	"os"
)

/*
根据csv地址文件，插入到地址库中
*/

func InsertAddressToDB(cfg *conf.InsertAddressConfig) error {
	if cfg.CsvPath == "" {
		return errors.New("csv path为空")
	}
	file, err := os.Open(cfg.CsvPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var (
		total, success, fail int
	)
	for {
		// Read返回的是一个数组，它已经帮我们分割了，
		record, err := reader.Read()
		total++
		// 如果读到文件的结尾，EOF的优先级比nil高！
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(cfg.CsvPath)
			log.Println(reader.FieldsPerRecord)
			return nil
		}
		address := record[0]
		addressInfo := new(po.AddressesInfo)
		addressInfo.Address = address
		addressInfo.Status = "used"
		addressInfo.UserID = cfg.UserId
		addressInfo.CoinType = cfg.CoinName
		//判断地址是否存在数据库
		if po.FindAddressIsExist(cfg.CoinName, address) {
			log.Printf("地址[%s]已经存在\n", address)
			continue
		}
		err = po.InsertAddressInfo(addressInfo)
		if err != nil {
			fail++
			log.Printf("插入地址[%s]错误： %v\n", address, err)
			continue
		}
		success++
	}
	fmt.Printf("插入地址成功，总共需要插入个数：%d,成功插入个数： %d，失败插入个数：%d", total, success, fail)
	return nil
}
