package stg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"rsksync/utils"
)

const TOKENSIZE = 16

type StgWSClient struct {
	conn     *websocket.Conn
	uri      string
	token    []byte
	response chan string
}

func NewStgWSClient(uri, token string) (*StgWSClient, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   "52.195.19.188:3145",
		Path:   "/",
	}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("failed to upgrade protocol to websocket")
	}

	tk, err := utils.Base64Decode([]byte(token))
	if err != nil {
		return nil, err
	}

	if len(tk) != TOKENSIZE {
		return nil, fmt.Errorf("token size isn't %d", TOKENSIZE)
	}

	c := &StgWSClient{
		conn:     conn,
		uri:      uri,
		token:    tk,
		response: make(chan string),
	}

	return c, nil
}

func (c *StgWSClient) handleMessage() {

	for {
		//res := new(Response)
		res, err := c.recvMsg()
		if err != nil {
			log.Println("go-xrp read ws response:", err)
			continue
		}
		log.Println("go read ws response:", string(res))
		c.response <- string(res)
	}
}

type BlockRequest struct {
	Epoch  uint64 `json:"epoch"`
	Offset uint32 `json:"offset"`
	Id     uint64 `json:"id"`
}

type HistoryRequest struct {
	Id        int64  `json:"id"`
	Type      string `json:"history_info"`
	AccountId int64  `json:"account_id"`
	Start     string `json:"starting_from"`
	Limit     uint64 `json:"limit"`
}

type Response map[string]interface{}

func (c *StgWSClient) sendMsg(data string) error {
	cipherdata, err := encrypt(c.token, data)
	if err != nil {
		return err
	}
	msg := utils.Base64Encode(cipherdata)
	log.Printf("send msg : %s", msg)
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *StgWSClient) recvMsg() ([]byte, error) {
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return nil, err
	}

	cipherdata, err := utils.Base64Decode(message)
	if err != nil {
		return nil, err
	}

	plaintdata, err := decrypt(c.token, cipherdata)
	if err != nil {
		return nil, err
	}

	log.Printf("msg : %s", string(plaintdata))
	return plaintdata, nil
}

func (c *StgWSClient) decrypt(src []byte, target interface{}) error {
	tmp, err := utils.Base64Decode(src)
	if err != nil {
		return err
	}

	data, err := decrypt(tmp, c.token)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

func encrypt(api_token []byte, payload string) ([]byte, error) {
	block, err := aes.NewCipher(api_token)
	if err != nil {
		return nil, err
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(payload))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(payload))
	return ciphertext, nil
}

func decrypt(api_token []byte, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("invalid ciphertext")
	}

	block, err := aes.NewCipher(api_token)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext[:], nil
}
