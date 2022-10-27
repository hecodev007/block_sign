package secure

import (
	"bnbsign/conf"
	"bnbsign/crypto"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RequestKMSAddr struct {
	Mch     string   `json:"mch"`
	Address []string `json:"address"`
}

type RequestKMSResult struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []ResponseKMSAddr `json:"data"`
}

type ResponseKMSAddr struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func GetPrivateKey(mch string, address string) (string, error) {
	logrus.Infof("[keyManager] GetPrivateKey address=%s", address)
	req := &RequestKMSAddr{Mch: mch, Address: []string{address}}
	ms, _ := json.Marshal(req)

	url := fmt.Sprintf("%s/%s", conf.Global.KMS.Url, "getKey")
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(ms))
	if err != nil {
		return "", err
	}
	timeout := time.Duration(conf.Global.KMS.HttpTimeout) * time.Second
	result, err := sendWithRetry(r, timeout, conf.Global.KMS.RetryCount)
	if err != nil {
		return "", err
	}

	response := &RequestKMSResult{}
	if err = json.Unmarshal(result, response); err != nil {
		return "", err
	}

	if response.Code != 0 {
		return "", errors.New(fmt.Sprintf("failed to get private key, message=%s", response.Message))
	}

	for _, d := range response.Data {
		if strings.ToLower(d.Address) == strings.ToLower(address) {
			priv, err := crypto.AesBase64Str(d.PrivateKey, conf.Global.Secret.TransportSecureKey, false)
			if err != nil {
				return "", err
			}
			return priv, nil
		}
	}
	return "", errors.New(fmt.Sprintf("failed to get private key"))
}

// sendWithRetry 支持重试的POST请求
// 重试次数+1 = 实际请求的次数
// 每次重试间隔+2秒（2、4、6、8、10...）
func sendWithRetry(r *http.Request, timeout time.Duration, retryCount uint32) ([]byte, error) {
	var (
		response []byte
		err      error
	)
	for i := 0; i < int(retryCount)+1; i++ {
		if i > 0 {
			sleepDuration := time.Duration((i + 1) * 2)
			fmt.Sprintf("%s [%d] retry in %d seconds", r.URL.RawPath, i+1, sleepDuration)
			time.Sleep(sleepDuration * time.Second)
		}

		response, err = send(r, timeout)
		if err == nil {
			return response, nil
		}
		fmt.Sprintf("HTTP send err %v", err)
	}
	return response, err
}

func send(r *http.Request, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}
