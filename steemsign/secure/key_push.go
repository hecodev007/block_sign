package secure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"steemsign/common/conf"
	"steemsign/common/log"
	"time"
)

type RequestPushKey struct {
	CoinCode string        `json:"coinCode"`
	Mch      string        `json:"mch"`
	Data     []PushKeyData `json:"data"`
}

type PushKeyData struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type KMSPushKeyResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func PushKeyToKMS(kmsReq *RequestPushKey) error {
	ms, _ := json.Marshal(kmsReq)
	url := fmt.Sprintf("%s%s", conf.Global.KMS.Url, "/pushKey")
	log.Infof("CALL KMS %s", url)
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(ms))
	if err != nil {
		return fmt.Errorf("failed to http.NewRequest KMS %v", err)
	}
	timeout := time.Duration(conf.Global.KMS.HttpTimeout) * time.Second
	result, err := sendWithRetry(r, timeout, conf.Global.KMS.RetryCount)
	if err != nil {
		return fmt.Errorf("failed to call [%s], KMS response: %v", url, err)
	}

	kmsResult := &KMSPushKeyResult{}
	if err = json.Unmarshal(result, kmsResult); err != nil {
		return fmt.Errorf("call KMS pushKey failed %v", err)
	}
	if kmsResult.Code != 0 {
		return fmt.Errorf("call KMS pushKey failed %s", kmsResult.Message)
	}
	return nil
}
