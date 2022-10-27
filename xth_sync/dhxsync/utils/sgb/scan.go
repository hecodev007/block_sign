package sgb

import (
	"bytes"
	"dhxsync/common/log"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type ScanTxResponse struct {
	Code int `json:"code"`
	Data struct {
		Transfer struct {
			Fee string `json:"fee"`
		} `json:"transfer"`
	} `json:"data"`
	Message string `json:"message"`
}

func GetFee(txhash string) string {
getFeetry:
	fee, err := getFee(txhash)
	if err != nil {
		log.Warn(err.Error() + " " + txhash)
		time.Sleep(time.Second * 3)
		goto getFeetry
	}
	return fee
}
func GetFee2(blockhash string, index int, txhash string) string {
	return ""
}
func getFee(txhash string) (fee string, err error) {
	params := []byte("{\"extrinsic_index\": \"\", \"hash\": \"" + txhash + "\"}")
	resp, err := http.Post("https://subgamescan.io/api/scan/extrinsic",
		"application/application/json",
		bytes.NewBuffer(params))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result := new(ScanTxResponse)
	err = json.Unmarshal(body, result)
	if err != nil {
		return
	}
	if result.Code != 0 {
		return fee, errors.New(result.Message)
	}
	return result.Data.Transfer.Fee, nil
}
