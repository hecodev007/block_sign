package bifrost

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type DotScanApi struct{
	Url string
	Key string
	Dec int //精度
}

func NewDotScanApi(url string,key string) *DotScanApi{
	return &DotScanApi{
		Url:url,
		Key:key,
		Dec:12,
	}
}

func (Api *DotScanApi)Post(path string,params interface{},ret interface{})error{
	header := "X-API-Key: "+Api.Key
	param_data,err := json.Marshal(params)
	if err != nil {
		return err
	}
	resp,err := http.Post(Api.Url+path,header,bytes.NewBuffer(param_data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body,ret)
}
type AccountInfoResponse struct{
	Code int64 `json:"code"`
	Message string `json:"message"`
	Data struct{
		Account struct{
			Balance decimal.Decimal `json:"balance"`
			Nonce int64 `json:"nonce"`
		} `json:"account"`
	} `json:"data"`
}
var NonceManage *NoncesStu
func init(){
	NonceManage = new(NoncesStu)
	NonceManage.data = make(map[string]nonceinfo)
}
type NoncesStu struct {
	expiry int64
	l sync.Mutex
	data map[string]nonceinfo
}
func (n *NoncesStu)Set(address string,nonce int64){
	n.l.Lock()
	defer n.l.Unlock()
	n.data[address] = nonceinfo{
		nonce:nonce,
		time:time.Now(),
	}
}
func (n *NoncesStu)Get(address string)(nonce int64,t time.Time){
	n.l.Lock()
	defer n.l.Unlock()
	v,ok := n.data[address]
	if ok {
		fmt.Println(address,v.nonce,v.time)
		return v.nonce,v.time
	}
	return 0,time.Time{}
}
type nonceinfo struct{
	nonce int64
	time time.Time
}

func (Api *DotScanApi)GetNonce(address string)(nonce uint64,err error){
	localnonce,t := NonceManage.Get(address)
	_,scan_nonce,err := Api.AccountInfo(address)
	if err != nil {
		return 0,err
	}
	fmt.Println("GetNonce",address,scan_nonce)
	if localnonce == 0 {
		return uint64(scan_nonce),err
	}
	n := 5
	for scan_nonce < localnonce{
		_,scan_nonce,err = Api.AccountInfo(address)
		if err != nil {
			return 0,err
		}
		n -= 1
		if n <= 0{
			break
		}
		time.Sleep(time.Second)
	}
	if scan_nonce >= localnonce{
		return uint64(scan_nonce),nil
	}

	now := time.Now()
	if now.Sub(t) > 180*time.Second{
		return uint64(scan_nonce),err
	}
	if err != nil {
		return 0, err
	}
	return uint64(localnonce),nil
}
func (Api *DotScanApi)AccountInfo(address string)(amount int64,nonce int64,err error){
	resp := new(AccountInfoResponse)
	//return 100000000000000,0, nil

	err = Api.Post("/api/scan/search",map[string]string{"key":address},resp)
	if err != nil {
		return 0,0, err
	}
	if resp.Code != 0 {

		return 0,0,errors.New(resp.Message)
	}
	return resp.Data.Account.Balance.Shift(12).IntPart(),resp.Data.Account.Nonce,nil
}
