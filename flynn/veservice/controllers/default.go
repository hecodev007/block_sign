package controllers

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vechain/thor/builtin"
	"io/ioutil"
	"math/big"
	"strings"
	"time"
	comm "veservice/common"
	"veservice/models"

	"github.com/astaxie/beego"

	//ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shopspring/decimal"
	//"github.com/vechain/thor/builtin/gen"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/tx"
)

// TokenABI is the input ABI used to generate the binding from.
const TokenABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"supply\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

var (
	EncryptWifMap map[string]string = make(map[string]string)
	WifKeyListMap map[string]string = make(map[string]string)
)

var (
	l         = big.NewInt(100)
	d         = new(big.Int).SetInt64(1000000000000000000)
	vthoLimit = new(big.Int).Mul(l, d)
)

type MainController struct {
	beego.Controller
}

func ReadUsbAConfig(usb_a string) {
	if usb_a == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_a)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	encryptList, _ := r2.ReadAll()
	beego.Debug("total load A file nums: ", len(encryptList))
	for i := 0; i < len(encryptList); i++ {
		//fmt.Println(encryptList[i][0], encryptList[i][1])
		address := strings.ToLower(encryptList[i][1])
		EncryptWifMap[address] = encryptList[i][0]
	}
}

func ReadUsbBConfig(usb_b string) {
	if usb_b == "" {
		return
	}

	cntb, err := ioutil.ReadFile(usb_b)
	if err != nil {
		panic(err.Error())
	}
	r2 := csv.NewReader(strings.NewReader(string(cntb)))
	keyList, _ := r2.ReadAll()
	beego.Debug("total load B file nums: ", len(keyList))
	for i := 0; i < len(keyList); i++ {
		//fmt.Println(keyList[i][0], keyList[i][1])
		address := strings.ToLower(keyList[i][1])
		WifKeyListMap[address] = keyList[i][0]
	}
}

func ReadConfig(usb_a string, usb_b string) {
	if usb_a == "" || usb_b == "" {
		return
	}

	ReadUsbAConfig(usb_a)
	ReadUsbBConfig(usb_b)
}

// VTHO = (1 + GasPriceCoef(128)/255) * Gas(21)
// "contractAddress":"0x0000000000000000000000000000456E65726779"
// 单位 amount: vet,vtho 10^18 = 1000000000000000000
// request json
//{
//	"data":[
//		{
//			"data":{
//				"from": "0x80a30A48BF3bE48b262DDAF6625dd55DbF4937F7",
//				"nonce":0,
//				"blockNumber":958720,
//				"contractAddress":"",
//				"tolist":[
//					{
//						"to":"0x5c29b47210a1999cbbd343ce105fc8704bf8f1d3",
//						"amount":"10000000000000000000"
//					}
//				]
//			}
//		}
//	]
//}
func (c *MainController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"result": nil,
		"error":  nil,
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		//beego.Trace(jsonObj)

		if jsonObj["data"] == nil {
			break
		}

		//beego.Debug(jsonObj["data"].(type))
		var hexs []*models.EthRawData
		switch v := jsonObj["data"].(type) {
		case []interface{}:
			list := jsonObj["data"].([]interface{})
			for i := 0; i < len(list); i++ {
				obj := list[i].(map[string]interface{})
				hex, _ := c.Sign(obj["data"].(map[string]interface{}))
				tx := new(models.EthRawData)
				tx.Hex = hex
				tx.Index = i
				hexs = append(hexs, tx)
			}
			break

		case map[string]interface{}:
			hex, _ := c.Sign(jsonObj["data"].(map[string]interface{}))
			tx := new(models.EthRawData)
			tx.Hex = hex
			tx.Index = 0
			hexs = append(hexs, tx)

		default:
			beego.Debug(v)
		}

		resp["result"] = hexs

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *MainController) Sign(obj map[string]interface{}) (string, error) {
	var ob models.EthRawTx
	ob.From = strings.ToLower(obj["from"].(string))
	coinName := obj["coinName"].(string)

	onlineAmount, vthoAmount, err := getAddressAmount(ob.From)

	if err != nil {
		return "", err
	}
	if vthoAmount.Cmp(vthoLimit) < 0 {
		return "", errors.New("vtho amount is less than 100")
	}
	ob.BlockNumber = int64(obj["blockNumber"].(float64))
	if ob.BlockNumber == 0 {
		resp, err := comm.GetJson(fmt.Sprintf("%s/blocks/best", beego.AppConfig.String("url")), nil)
		if err != nil {
			return "", fmt.Errorf("get latest block number error,err=%v", err)
		}
		var block map[string]interface{}
		err1 := json.Unmarshal(resp, &block)
		if err1 != nil || block == nil {
			return "", err1
		}
		blockNumber := block["number"].(float64)
		ob.BlockNumber = int64(blockNumber) + 10 //延迟5个区块上账
	}

	ob.Nonce = uint64(obj["nonce"].(float64))
	if ob.Nonce == 0 {
		ob.Nonce = uint64(time.Now().UnixNano())
	}
	//ob.To = obj["to"].(string)
	//ob.GasLimit = uint64(obj["gasLimit"].(float64))
	//ob.GasPrice = int64(obj["gasPrice"].(float64))

	if obj["data"] != nil {
		ob.Data = obj["data"].(string)
	}

	var trx *tx.Transaction
	chainTag := beego.AppConfig.DefaultString("chaintag", "0x4a")
	tag, _ := hex.DecodeString(removeHex0x(chainTag))
	gas := beego.AppConfig.DefaultInt64("gas", 60000)
	if obj["contractAddress"] != nil && len(obj["contractAddress"].(string)) > 0 {
		if obj["data"] == nil || obj["data"].(string) == "" {

			method, found := builtin.Energy.ABI.MethodByName("transfer")
			if !found {
				//beego.Debug("method not found")
				return "", errors.New("transfer method not found")
			}

			// VTHO = (1 + GasPriceCoef(128)/255) * Gas(21)
			contractAddress, _ := thor.ParseAddress(obj["contractAddress"].(string))

			builder := new(tx.Builder).ChainTag(tag[0]).
				BlockRef(tx.NewBlockRef(uint32(ob.BlockNumber))).
				Expiration(720).
				GasPriceCoef(0).
				Gas(uint64(gas) * uint64(len(obj["tolist"].([]interface{})))).
				DependsOn(nil).
				Nonce(ob.Nonce)

			//builder := new(tx.Builder).ChainTag(0xa4).
			//	BlockRef(tx.BlockRef{0, 0, 0, 0, 0xaa, 0xbb, 0xcc, 0xdd}).
			//	Expiration(32).
			//	GasPriceCoef(128).
			//	Gas(210000).
			//	DependsOn(nil).
			//	Nonce(12345678)

			//builder := new(tx.Builder).ChainTag(74).
			//	BlockRef(tx.NewBlockRef(uint32(ob.BlockNumber))).
			//	Expiration(720).
			//	GasPriceCoef(128).
			//	Gas(51000 * uint64(len(obj["tolist"].([]interface{})))).
			//	DependsOn(nil).
			//	Nonce(ob.Nonce)

			// more to list
			if obj["tolist"] != nil {
				list := obj["tolist"].([]interface{})
				for i := 0; i < len(list); i++ {
					_obj := list[i].(map[string]interface{})
					_tmpamount, err := decimal.NewFromString(_obj["amount"].(string))
					if err != nil {
						beego.Debug("convert amount fail !")
						return "", errors.New("convert amount fail !")
					}
					amount := _tmpamount.Coefficient()
					if vthoAmount.Cmp(amount) < 0 {
						return "", fmt.Errorf("online amount is less than transfer amount,OnlineAmount=[%d],TransAmount=[%d]", onlineAmount.Int64(), amount.Int64())
					}
					to, _ := thor.ParseAddress(_obj["to"].(string))
					packed, err := method.EncodeInput(&to, amount)
					if packed == nil || err != nil {
						beego.Debug(err)
						return "", err
					}

					//builder = builder.Clause(tx.NewClause(&contractAddress).WithValue(big.NewInt(0)).WithData(packed))
					builder = builder.Clause(tx.NewClause(&contractAddress).WithData(packed))
				}
			} else {
				beego.Debug("tolist not found")
				return "", errors.New("tolist not found")
			}
			beego.Info(ob.BlockNumber)
			beego.Info("代币交易")
			trx = builder.Build()
		}

	} else if coinName == "vet" {
		builder := new(tx.Builder).ChainTag(tag[0]).
			BlockRef(tx.NewBlockRef(uint32(ob.BlockNumber))).
			Expiration(720).
			GasPriceCoef(128).
			Gas(uint64(gas) * uint64(len(obj["tolist"].([]interface{})))).
			DependsOn(nil).
			Nonce(ob.Nonce)
		if obj["tolist"] != nil {
			list := obj["tolist"].([]interface{})
			for i := 0; i < len(list); i++ {

				_obj := list[i].(map[string]interface{})
				_tmpamount, err := decimal.NewFromString(_obj["amount"].(string))
				if err != nil {
					beego.Debug("convert amount fail !")
					return "", errors.New("convert amount fail !")
				}
				amount := _tmpamount.Coefficient()
				if onlineAmount.Cmp(amount) < 0 {
					return "", fmt.Errorf("online amount is less than transfer amount,OnlineAmount=[%d],TransAmount=[%d]", onlineAmount.Int64(), amount.Int64())
				}
				to, _ := thor.ParseAddress(_obj["to"].(string))
				builder = builder.Clause(tx.NewClause(&to).WithValue(amount).WithData([]byte(ob.Data)))
			}
		} else {
			beego.Debug("tolist not found")
			return "", errors.New("tolist not found")
		}
		beego.Info("主链币交易")
		trx = builder.Build()
	} else {
		return "", errors.New("***unknown coin name transfer***")
	}

	from := strings.ToLower(ob.From)
	if EncryptWifMap[from] == "" || WifKeyListMap[from] == "" {
		beego.Debug("key not found !", from)
		return "", fmt.Errorf("key not found,address=%s", from)
	}

	wif, _ := comm.AesDecrypt(EncryptWifMap[from], []byte(WifKeyListMap[from]))
	if wif == "" {
		beego.Debug("decrypt key fail !", from)
		return "", fmt.Errorf("decrypt key fail !,address=%s", from)
	}

	privKey, _ := crypto.HexToECDSA(wif)
	if privKey == nil {
		return "", errors.New("private key is null")
	}

	// 签名
	sig, signerr := crypto.Sign(trx.SigningHash().Bytes(), privKey)
	if signerr != nil {
		beego.Debug("could not sign transaction", signerr)
		return "", fmt.Errorf("could not sign transaction,err=%v", signerr)
	}

	trx = trx.WithSignature(sig)
	signdata, _ := rlp.EncodeToBytes(trx)
	hextx := hex.EncodeToString(signdata)
	if !strings.HasPrefix(hextx, "0x") {
		hextx = "0x" + hextx
	}

	return hextx, nil
}

func removeHex0x(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		return hexStr[2:]
	}
	return hexStr
}

func getAddressAmount(address string) (*big.Int, *big.Int, error) {
	resp, err := comm.GetJson(fmt.Sprintf("%s/accounts/%s", beego.AppConfig.String("url"), address), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get address=[%s] amount error,err=%v", address, err)
	}
	var data map[string]interface{}
	err1 := json.Unmarshal(resp, &data)
	if err1 != nil || data == nil {
		return nil, nil, err1
	}
	balance := data["balance"].(string)
	energy := data["energy"].(string)
	amount, isOK := new(big.Int).SetString(removeHex0x(balance), 16)
	if !isOK {
		return nil, nil, errors.New("balance convert to big.Int error")
	}
	vtho, isOk2 := new(big.Int).SetString(removeHex0x(energy), 16)
	if !isOk2 {
		return nil, nil, errors.New("energy convert to big.Int error")
	}
	return amount, vtho, nil
}
