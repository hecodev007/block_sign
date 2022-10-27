package main

import (
	"encoding/csv"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"io"
	"os"
	"runtime"
)

func InitDB() {
	maxCPU := runtime.NumCPU()
	syncdsn := beego.AppConfig.String("syncdsn")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", syncdsn)
	orm.SetMaxIdleConns("default", maxCPU*2)
	orm.SetMaxOpenConns("default", maxCPU*4)
}

func GetAmount(fromAddr string) (error, float64) {
	// 查找from金额
	totalAmount := 0.0

	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select sum(vout_value) as sumvalue from block_tx_vout where vout_address = ? and status = 1", fromAddr).Values(&maps)
	if err == nil && nums > 0 {
		if maps[0]["sumvalue"] == nil {
			totalAmount = 0
		} else {
			totalAmount = common.StrToFloat64(maps[0]["sumvalue"].(string))
		}
	}

	return nil, totalAmount
}

func main() {
	var path string = beego.AppConfig.String("csvdir")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	fs, err := os.Open(path)
	if err != nil {
		log.Debug("can not open the file, err is %+v", err)
		return
	}
	defer fs.Close()

	InitDB()

	list := [][]string{}
	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Debug("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}

		list = append(list, row)
	}

	clientsFile, err := os.OpenFile("./check.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Debug(err)
		return
	}
	defer clientsFile.Close()
	clients := csv.NewWriter(clientsFile)
	for i := 0; i < len(list); {
		err, amount := GetAmount(list[i][0])
		if err != nil {
			log.Debug(err, list[i][0])
			continue
		}

		val := decimal.NewFromFloat(amount).Div(decimal.New(1, int32(beego.AppConfig.DefaultInt("precision", 8))))
		list[i] = append(list[i], val.String())
		clients.Write(list[i])
		i++
	}
	clients.Flush()
}
