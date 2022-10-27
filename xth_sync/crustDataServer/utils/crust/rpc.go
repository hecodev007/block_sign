package crust

import (
	"crustDataServer/common/log"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type request struct {
	ID      int64         `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
type response struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}
type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"` /* required */
}

func (e *Error) Error() string {
	return e.Message
}

type RpcClient struct {
	*websocket.Conn
	id        int64
	host      string
	lock      sync.Mutex
	receiveCh sync.Map
}

func NewRpcClient(host string, K, S string) (r *RpcClient) {
	r = &RpcClient{
		host: host,
	}
	err := r.dail()
	if err != nil {
		panic(err.Error())
	}
	go r.read()
	return r
}
func (rpc *RpcClient) read() {
	go func() {
		for {
			messageType, message, err := rpc.Conn.ReadMessage()
			if err != nil {
				log.Info(err.Error())
				coin := rpc.Conn
				//重连
				if err = rpc.dail(); err != nil {
					log.Info(err.Error())
				}
				coin.Close()
				continue
			}
			rpc.onMesssge(messageType, message)
		}
	}()
}

func (rpc *RpcClient) dail() error {
	c, _, err := websocket.DefaultDialer.Dial(rpc.host, nil)
	if err != nil {
		return err
	}
	rpc.Conn = c
	return nil
}
func (rpc *RpcClient) Call(method string, params ...interface{}) (ret []byte, err error) {
	//
	rpc.lock.Lock()
	rpc.id++
	request := request{
		ID:      rpc.id,
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}
	ch := make(chan []byte, 1)
	rpc.receiveCh.Store(rpc.id, ch)
	defer rpc.receiveCh.Delete(request.ID)
	defer close(ch)
	body, _ := json.Marshal(request)
	err = rpc.Conn.WriteMessage(websocket.TextMessage, body)
	rpc.lock.Unlock()
	////

	if err != nil {
		return nil, err
	}
	var result []byte
	select {
	case result = <-ch: //
	case <-time.After(10 * time.Second): //超时10s
		err = errors.New("receive timeout")
		return nil, err
	}
	res := new(response)
	err = json.Unmarshal(result, res)
	if err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return res.Result, nil
}

func (rpc *RpcClient) onMesssge(mtype int, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("onMessage error:", r)
		}
	}()
	id := gjson.Get(string(data), "id").Int()
	value, ok := rpc.receiveCh.Load(id)
	if !ok {
		return
	}

	ch, ok := value.(chan []byte)
	if !ok {
		panic("")
	}

	select {
	case ch <- data:
	default:
	}

}
