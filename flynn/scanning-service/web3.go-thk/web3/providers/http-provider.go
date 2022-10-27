package providers

import (
	"bytes"
	"encoding/json"
	"github.com/group-coldwallet/scanning-service/web3.go-thk/web3/providers/util"
	"io/ioutil"
	"net/http"
	"time"
)

type HTTPProvider struct {
	address string
	timeout int32
	secure  bool
	client  *http.Client
}

func NewHTTPProvider(address string, timeout int32, secure bool) *HTTPProvider {
	return newHTTPProviderWithClient(address, timeout, secure, &http.Client{
		Timeout: time.Second * time.Duration(timeout),
	})
}

func newHTTPProviderWithClient(address string, timeout int32, secure bool, client *http.Client) *HTTPProvider {
	provider := new(HTTPProvider)
	provider.address = address
	provider.timeout = timeout
	provider.secure = secure
	provider.client = client

	return provider
}

func (provider HTTPProvider) SendRequest(v interface{}, method string, params interface{}) error {

	bodyString := util.JsonParam{Method: method, Params: params}

	prefix := "http://"
	if provider.secure {
		prefix = "https://"
	}
	bufferparams, err := json.Marshal(bodyString)
	// body := strings.NewReader(bodyString.AsJsonString())
	req, err := http.NewRequest("POST", prefix+provider.address, bytes.NewBuffer(bufferparams))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := provider.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var bodyBytes []byte

	if resp.StatusCode == 200 {
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		} else {
			return json.Unmarshal(bodyBytes, v)
		}
	}

	return json.Unmarshal(bodyBytes, v)

}

func (provider HTTPProvider) Close() error {
	return nil
}
