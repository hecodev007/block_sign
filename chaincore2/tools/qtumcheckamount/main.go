package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	"github.com/shopspring/decimal"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//var _abi abi.ABI

func init() {
	maxCPU := runtime.NumCPU()
	syncdsn := beego.AppConfig.String("syncdsn")
	userdsn := beego.AppConfig.String("userdsn")
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", syncdsn)
	orm.RegisterDataBase("user","mysql", userdsn)
	orm.SetMaxIdleConns("default", maxCPU*2)
	orm.SetMaxOpenConns("default", maxCPU*4)
	ormdebug, _ := beego.AppConfig.Bool("ormdebug")
	orm.Debug = ormdebug

	// 注册model模型
	//orm.RegisterModel(new(models.User))
	//调用 RunCommand 执行 orm 命令。
	//orm.RunCommand()
}

func Get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		beego.Debug("qtum: %v ", err)
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		beego.Debug("qtum: %v ", err)
		return nil, err
	}

	return content, nil
}

func Request(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("nodeurl")).SetTimeout(time.Second*5, time.Second*10)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	if beego.AppConfig.DefaultBool("enabletls", false) {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)
	return req.Bytes()
}

func GetTimeHeight(starttime string, stoptime string) (int64, int64) {
	startheight := int64(0)
	stopheight := int64(0)

	{
		o := orm.NewOrm()
		var maps []orm.Params
		nums, err := o.Raw("select min(height) as height from block_info where time >= ? and time < ?", starttime, stoptime).Values(&maps)
		if err == nil && nums > 0 && len(maps) > 0 {
			startheight = StrToInt64(maps[0]["height"].(string))
		}
	}

	{
		o := orm.NewOrm()
		var maps []orm.Params
		nums, err := o.Raw("select max(height) as height from block_info where time >= ? and time < ?", starttime, stoptime).Values(&maps)
		if err == nil && nums > 0 && len(maps) > 0 {
			stopheight = StrToInt64(maps[0]["height"].(string))
		}
	}

	return  startheight, stopheight
}

func GetAmount(fromAddress string, startheight int64, stopheight int64) (error, float64) {
	// 查找from金额
	totalAmount := 0.0

	// 校验合约余额
	//if false {
	//	_from := b58addr.ToHexString(fromAddress)
	//	packed, err := _abi.Pack("balanceOf", common.HexToAddress(_from))
	//	if err != nil {
	//		beego.Debug(err)
	//		return err, totalAmount
	//	}
	//	hextx := common.Bytes2Hex(packed)
	//	respdata, err := Request("callcontract", []interface{}{"f2033ede578e17fa6231047265010445bca8cf1c", hextx})
	//	if err != nil {
	//		return err, totalAmount
	//	}
	//	var rawdata map[string]interface{}
	//	err = json.Unmarshal(respdata, &rawdata)
	//	if err != nil {
	//		return err, totalAmount
	//	}
	//	//beego.Debug(rawdata)
	//	result := rawdata["result"].(map[string]interface{})
	//	executionResult := result["executionResult"].(map[string]interface{})
	//	qrc20amount := StrBaseToInt64(executionResult["output"].(string), 16)
	//	totalAmount, _ = decimal.NewFromFloat(float64(qrc20amount)).Div(decimal.New(1, 8)).Float64()
	//}

	{
		o := orm.NewOrm()
		var maps []orm.Params
		nums, err := o.Raw("select sum(amount) as total from contract_tx where height >= ? and height <= ? and toaddress = ? and contract_address = ?", startheight, stopheight, fromAddress, "f2033ede578e17fa6231047265010445bca8cf1c").Values(&maps)
		if err == nil && nums > 0 && len(maps) > 0 {
			if maps[0]["total"] != nil {
				qrc20amount, _ := decimal.NewFromString(maps[0]["total"].(string))
				totalAmount, _ = qrc20amount.Div(decimal.New(1, 8)).Float64()
			}
		}
	}

	return nil, totalAmount
}

func main() {
	var path string = beego.AppConfig.String("csvdir")
	var writepath string = "qc_check.csv"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if len(os.Args) > 2 {
		writepath = os.Args[2]
	}

	fs, err := os.Open(path)
	if err != nil {
		beego.Debug("can not open the file, err is %+v", err)
	}
	defer fs.Close()

	//_abi, err = abi.JSON(strings.NewReader(TokenABI))
	startheight, stopheight := GetTimeHeight(beego.AppConfig.String("starttime"), beego.AppConfig.String("stoptime"))
	beego.Debug(startheight, stopheight)

	list := [][]string {}
	r := csv.NewReader(fs)
	//针对大文件，一行一行的读取文件
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			beego.Debug("can not read, err is %+v", err)
		}
		if err == io.EOF {
			break
		}

		list = append(list, row)
	}

	clientsFile, err := os.OpenFile(writepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		beego.Debug(err)
		return
	}
	defer clientsFile.Close()
	clients := csv.NewWriter(clientsFile)
	for i := 0; i < len(list); {
		if i == 0 {
			i++
			continue
		}
		beego.Debug(i, list[i][8])
		err, amount := GetAmount(list[i][8], startheight, stopheight)
		if err != nil {
			beego.Debug(err, list[i][0])
			continue
		}

		list[i] = append(list[i], fmt.Sprintf("%.8f", amount))
		clients.Write(list[i])
		i++
	}
	clients.Flush()
}
