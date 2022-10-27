package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	client IRpcClient
}
type IRpcClient interface {
	SendRequest(method string, result interface{}, params interface{}) error
}

func NewRpcClient(url, user, password string) (*Client, error) {
	r := new(Client)
	var ic IRpcClient
	if strings.HasPrefix(url, "ws") || strings.HasPrefix(url, "wss") {
		// 连接websocket
		socket := NewWebsocket(url)
		ic = &socket
		//return client, errors.New("do not support websocket")
	} else if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "https") {
		ic = Dial(url, user, password)
	} else {
		return nil, fmt.Errorf("unsopport url  %s", url)
	}
	r.client = ic
	return r, nil
}

func (r *Client) Post(method string, result interface{}, params interface{}) error {
	return r.client.SendRequest(method, result, params)
}

//http
type RpcClient struct {
	rpcUrl      string
	rpcUser     string
	rpcPassword string
}

type RequestBody struct {
	ReqNotHaveParams
	Params interface{} `json:"params"`
}
type ReqNotHaveParams struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Id      int    `json:"id"`
}
type RespBody struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
}
type RespErrorBody struct {
	JsonRpc string                 `json:"jsonrpc"`
	Error   map[string]interface{} `json:"error"`
	Id      int                    `json:"id"`
}

//初始化一个rpc客户端
func Dial(url, user, password string) *RpcClient {
	return &RpcClient{
		rpcUrl:      url,
		rpcUser:     user,
		rpcPassword: password,
	}
}

func (rpc *RpcClient) SendRequest(method string, result interface{}, params interface{}) error {
	id := rand.Intn(10000)
	var (
		reqBytes []byte
		err      error
	)
	if params != nil {
		var reqBody RequestBody
		reqBody.JsonRpc = "2.0"
		reqBody.Id = id
		reqBody.Method = method
		reqBody.Params = params
		reqBytes, err = json.Marshal(reqBody)
	} else {
		var reqBody ReqNotHaveParams
		reqBody.JsonRpc = "2.0"
		reqBody.Id = id
		reqBody.Method = method
		reqBytes, err = json.Marshal(reqBody)
	}
	if err != nil {
		return err
	}
	reqBuf := bytes.NewBuffer(reqBytes)
	var (
		req *http.Request
	)
	if req, err = http.NewRequest(http.MethodPost, rpc.rpcUrl, reqBuf); err != nil {
		return fmt.Errorf("http send error ,%v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	//设置rpc的用户和密码
	//如果为空就不设置
	if rpc.rpcUser != "" && rpc.rpcPassword != "" {
		req.SetBasicAuth(rpc.rpcUser, rpc.rpcPassword)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client do error, %v", err)
	}
	defer res.Body.Close()

	resp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("io read error, %v", err)
	}
	//解析resp
	var response RespBody
	if err := json.Unmarshal(resp, &response); err != nil {
		return fmt.Errorf("parse resp error,Err=【%v】", err)
	}
	if response.Result == nil {
		return fmt.Errorf("unknown error, %s", string(resp))
	}
	data, err := json.Marshal(response.Result)
	if err != nil {
		return fmt.Errorf("rpc marshal result error,Err=【%v】,Data=[%s]", err, string(resp))
	}
	//fmt.Println(string(data))
	err = json.Unmarshal(data, result)
	if err != nil {
		return fmt.Errorf("unmarshal result error, %v", err)
	}
	return nil
}

//websocket

type Socket struct {
	Conn              *websocket.Conn
	WebsocketDialer   *websocket.Dialer
	Url               string
	ConnectionOptions ConnectionOptions
	RequestHeader     http.Header
	OnConnected       func(socket Socket)
	OnTextMessage     func(message string, socket Socket)
	OnBinaryMessage   func(data []byte, socket Socket)
	OnConnectError    func(err error, socket Socket)
	OnDisconnected    func(err error, socket Socket)
	OnPingReceived    func(data string, socket Socket)
	OnPongReceived    func(data string, socket Socket)
	IsConnected       bool
	sendMu            *sync.Mutex // Prevent "concurrent write to websocket connection"
	receiveMu         *sync.Mutex
	reConnectNum      int
}

type ConnectionOptions struct {
	UseCompression bool
	UseSSL         bool
	Proxy          func(*http.Request) (*url.URL, error)
	Subprotocols   []string
}

func NewWebsocket(url string) Socket {
	return Socket{
		Url:           url,
		RequestHeader: http.Header{},
		ConnectionOptions: ConnectionOptions{
			UseCompression: false,
			UseSSL:         false,
		},
		WebsocketDialer: &websocket.Dialer{},
		sendMu:          &sync.Mutex{},
		receiveMu:       &sync.Mutex{},
	}
}

func (socket *Socket) setConnectionOptions() {
	socket.WebsocketDialer.EnableCompression = socket.ConnectionOptions.UseCompression
	socket.WebsocketDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: socket.ConnectionOptions.UseSSL}
	socket.WebsocketDialer.Proxy = socket.ConnectionOptions.Proxy
	socket.WebsocketDialer.Subprotocols = socket.ConnectionOptions.Subprotocols
}

func (socket *Socket) Connect() {
	var err error
	socket.setConnectionOptions()

	socket.Conn, _, err = socket.WebsocketDialer.Dial(socket.Url, socket.RequestHeader)

	if err != nil {
		log.Println("Error while connecting to server ", err)
		socket.IsConnected = false
		if socket.OnConnectError != nil {
			socket.OnConnectError(err, *socket)
		}
		return
	}

	log.Println("Connected to server")

	if socket.OnConnected != nil {
		socket.IsConnected = true
		socket.OnConnected(*socket)
	}

	defaultPingHandler := socket.Conn.PingHandler()
	socket.Conn.SetPingHandler(func(appData string) error {
		log.Println("Received PING from server")
		if socket.OnPingReceived != nil {
			socket.OnPingReceived(appData, *socket)
		}
		return defaultPingHandler(appData)
	})

	defaultPongHandler := socket.Conn.PongHandler()
	socket.Conn.SetPongHandler(func(appData string) error {
		log.Println("Received PONG from server")
		if socket.OnPongReceived != nil {
			socket.OnPongReceived(appData, *socket)
		}
		return defaultPongHandler(appData)
	})

	defaultCloseHandler := socket.Conn.CloseHandler()
	socket.Conn.SetCloseHandler(func(code int, text string) error {
		result := defaultCloseHandler(code, text)
		log.Println("Disconnected from server ", result)
		if socket.OnDisconnected != nil {
			socket.IsConnected = false
			socket.OnDisconnected(errors.New(text), *socket)
		}
		return result
	})

	go func() {
		for {
			socket.receiveMu.Lock()
			messageType, message, err := socket.Conn.ReadMessage()
			socket.receiveMu.Unlock()
			if err != nil {
				log.Println("read:", err)
			}
			switch messageType {
			case websocket.TextMessage:
				if socket.OnTextMessage != nil {
					socket.OnTextMessage(string(message), *socket)
				}
			case websocket.BinaryMessage:
				if socket.OnBinaryMessage != nil {
					socket.OnBinaryMessage(message, *socket)
				}
			}
		}
	}()
}

func (socket *Socket) SendText(message string) {
	err := socket.send(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("write:", err)
		return
	}
	log.Println("发送数据： ", message)
}

func (socket *Socket) SendBinary(data []byte) {
	err := socket.send(websocket.BinaryMessage, data)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (socket *Socket) send(messageType int, data []byte) error {
	socket.sendMu.Lock()
	err := socket.Conn.WriteMessage(messageType, data)
	socket.sendMu.Unlock()
	return err
}

func (socket *Socket) Close() {
	err := socket.send(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
	}
	socket.Conn.Close()
	if socket.OnDisconnected != nil {
		socket.IsConnected = false
		socket.OnDisconnected(err, *socket)
	}
}

type JsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Id      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

func (socket *Socket) SendRequest(method string, result interface{}, params interface{}) error {
	err := socket.reConnect()
	if err != nil {
		return err
	}
	reqData := map[string]interface{}{
		"id":      time.Now().Unix(),
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	dd, _ := json.Marshal(reqData)
	socket.sendMu.Lock()
	err = socket.Conn.WriteMessage(websocket.BinaryMessage, dd)
	socket.sendMu.Unlock()
	if err != nil {
		return fmt.Errorf("ws send req data error,Err=%v", err)
	}
	socket.receiveMu.Lock()
	_, respData, err := socket.Conn.ReadMessage()
	socket.receiveMu.Unlock()
	if err != nil {
		return fmt.Errorf("ws resp data error,Err=%v", err)
	}
	if len(respData) == 0 {
		return errors.New("ws resp data is null")
	}
	var resp JsonRpcResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return fmt.Errorf("ws json unmarshal resp data error,err=%v", err)
	}
	if resp.Error != nil {
		errData, _ := json.Marshal(resp.Error)
		return fmt.Errorf("ws resp error ,err=%s", string(errData))
	}
	if resp.Result == nil {
		return errors.New("ws resp result is null")
	}
	res, _ := json.Marshal(resp.Result)
	err = json.Unmarshal(res, result)
	if err != nil {
		return err
	}
	return nil
}

func (socket *Socket) reConnect() error {
	if socket.Conn == nil {
		var err error
		socket.setConnectionOptions()

		socket.Conn, _, err = socket.WebsocketDialer.Dial(socket.Url, socket.RequestHeader)

		if err != nil {
			log.Printf("error while connecting to server,err=%v,reConnect num is %d ", err, socket.reConnectNum)
			socket.reConnectNum++
			socket.IsConnected = false
			if socket.OnConnectError != nil {
				socket.OnConnectError(err, *socket)
			}
			if socket.reConnectNum >= 3 {
				return err
			}
			return socket.reConnect()
		}
		socket.reConnectNum = 0
	}

	return nil
}
