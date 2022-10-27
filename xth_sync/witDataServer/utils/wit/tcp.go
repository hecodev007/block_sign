package wit

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type Conn struct {
	chs  *SyncMap
	url  string
	conn *net.TCPConn
}

func NewConn(addr string) *Conn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err.Error())
	}
	ret := &Conn{
		url:  addr,
		conn: conn,
		chs:  NewSyncMap(),
	}
	go ret.Start()
	return ret
}

func (c *Conn) Start() {
	var msg []byte
	for {
		buf, err := c.Read()
		if err != nil {
			panic(err.Error())
			return
		}
		//log.Println("tcp", strings.HasSuffix(string(buf), "}\n"), string(buf))
		if strings.Contains(string(buf), "\n") {
			bufs := strings.Split(string(buf), "\n")
			msg = append(msg, []byte(bufs[0])...)
			id := gjson.Get(string(msg), "id").Int()
			ch := c.chs.Load(id)
			if ch == nil {
				continue
			} else {
				ch <- msg
			}
			for i := 1; i < len(bufs)-1; i++ {
				ch <- []byte(bufs[i])
			}
			msg = []byte(bufs[len(bufs)-1])
		} else {
			msg = append(msg, buf...)
		}
	}
}
func (c *Conn) Write(msg []byte) error {
	//println(string(msg))
	_, err := c.conn.Write(msg)
	return err
}
func (c *Conn) Connect() error {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", c.url)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}
func (c *Conn) Read() ([]byte, error) {
	var buf [4096]byte
	n, err := c.conn.Read(buf[0:])
	if err != nil {
		return nil, err
	}
	return buf[0:n], nil
}
func (c *Conn) Call(id int64, param interface{}) ([]byte, error) {
	request, _ := json.Marshal(param)
	request = []byte(string(request) + "\n")
	err := c.Write(request)
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte)
	c.chs.Store(id, ch)
	defer func() {
		c.chs.Del(id)
	}()
	var ret []byte
	select {
	case ret = <-ch:
		//println(string(ret))
		return ret, nil
	case <-time.After(time.Second * 300):
		return nil, errors.New("request time out")
	}
}
