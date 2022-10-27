package test

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/test/compiler"
	"io/ioutil"
	"strings"
)

type ReqBody struct {
	Method string
	Params map[string]interface{}
}
type Account struct {
	ChainId  string `json:"chainId"`
	Contract string `json:"contract"`
}
type Accounts struct {
	Method string  `json:"method"`
	Params Account `json:"params"`
}

func RpcCompileContract(name string) (res map[string]interface{}, err error) {
	data := new(map[string]interface{})
	contents, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	//var param Accounts
	//param.Method = "CompileContract"
	//param.Params.Contract = string(contents)
	//data1, err := json.Marshal(param)
	//resp, err := http.Post("http://192.168.1.13:8091",
	//	"application/json;charset=UTF-8",
	//	bytes.NewBuffer(data1))
	//defer resp.Body.Close()
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//str := string(body)
	//println(str)

	result := string(contents)
	ress, err := compiler.CompileSolidityString("", result)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", ress)
	for key, v := range ress {
		if strings.Contains(key, "<stdin>:") {
			newKey := strings.Replace(key, "<stdin>:", "", -1)
			delete(ress, key)
			ress[newKey] = v
		}
	}
	solcres, err := json.Marshal(ress)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(solcres, data); err != nil {
		return nil, err
	}
	return *data, err
}
func CompileContract(name ...string) (res map[string]interface{}, err error) {
	data := new(map[string]interface{})
	ress, err := compiler.CompileSolidity("", name...)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("%v\n", ress)
	for key, v := range ress {
		if strings.Contains(key, "<stdin>:") {
			newKey := strings.Replace(key, "<stdin>:", "", -1)
			delete(ress, key)
			ress[newKey] = v
		}
	}
	solcres, err := json.Marshal(ress)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(solcres, data); err != nil {
		return nil, err
	}
	return *data, err
}
