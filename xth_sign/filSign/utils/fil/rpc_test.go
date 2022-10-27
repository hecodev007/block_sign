package fil

import (
	"errors"
	"filSign/common/log"
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func Test_rpc(t *testing.T) {
	client := NewRpcClient("https://api.node.glif.io/rpc/v0", "", "")
	client = NewRpcClient("http://54.150.109.55:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.cgeQ3qhAlUgOaGzJ6BkhOl0mpTfjS7AzYjnG85IoXcI", "", "")
	client = NewRpcClient("http://13.231.123.210:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.h7ltPKwr-nWCqQih0c0rCiQ5R9TXpg20bkY2feKuD_0", "", "")
	n := 1
	for {
		st := time.Now()

		fee, err := client.BaseFee()
		//t.Log(fee, err)
		if err != nil {
			t.Log(err.Error(), time.Since(st).String())
			return
		}
		nonce, err := client.GetNonce("f1deqws4ivpx3jjpcnub4zf7333h3vk35jueh4w4q")
		if err != nil {
			t.Log(err.Error(), time.Since(st).String())
			return
		}
		t.Log(nonce, time.Since(st).String(), n)
		if time.Since(st).Seconds() > 15 {
			return
		}
		value, err := client.GetBalance("f1deqws4ivpx3jjpcnub4zf7333h3vk35jueh4w4q")
		if err != nil {
			t.Log(err.Error(), time.Since(st).String())
			return
		}
		t.Log(fee, nonce, value.String(), time.Since(st).String(), n)
		n++
		time.Sleep(time.Second)
	}
}

func Test_rpsc(t *testing.T) {
	client := NewRpcClient("http://54.150.109.55:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.cgeQ3qhAlUgOaGzJ6BkhOl0mpTfjS7AzYjnG85IoXcI", "", "")
	client1 := NewRpcClient("http://54.150.109.55:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.cgeQ3qhAlUgOaGzJ6BkhOl0mpTfjS7AzYjnG85IoXcI", "", "")
	client2 := NewRpcClient("http://13.231.123.210:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.h7ltPKwr-nWCqQih0c0rCiQ5R9TXpg20bkY2feKuD_0", "", "")

	//client2 := NewRpcClient("http://54.150.109.55:1234/rpc/v0?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.cgeQ3qhAlUgOaGzJ6BkhOl0mpTfjS7AzYjnG85IoXcI", "", "")
	//client = NewRpcClient("https://api.node.glif.io/rpc/v0", "", "")
	//client1 = NewRpcClient("https://api.node.glif.io/rpc/v0", "", "")
	// = NewRpcClient("https://api.node.glif.io/rpc/v0", "", "")
	rpcs := []*RpcClient{client, client1, client2}
	n := 1
	for {
		st := time.Now()
		nonce, err := GetNonce(rpcs, "f1deqws4ivpx3jjpcnub4zf7333h3vk35jueh4w4q")
		if err != nil {
			t.Log(err.Error(), time.Since(st).String())
			break
		}
		t.Log(nonce, n, time.Since(st).Seconds())
		n++
	}
	time.Sleep(time.Second * 600)
}
func GetNonce(rpcs []*RpcClient, addr string) (n int64, err error) {
	ch := make(chan int64)
	defer func() {
		close(ch)
	}()
	for _, cli := range rpcs {
		go func(r *RpcClient) {
			defer func() {
				if err := recover(); err != nil {
				}
			}()
			nonce, err := r.GetNonce(addr)
			if err != nil {
				fmt.Println(err.Error())
				log.Info(err.Error())
				return
			}
			ch <- nonce
		}(cli)
	}
	select {
	case n = <-ch:
		return n, nil
	case <-time.After(time.Second * 10):
		return 0, errors.New("获取nonce超时")
	}
}
func BaseFee(rpcs []*RpcClient) (n int64, err error) {
	ch := make(chan int64)
	defer func() {
		close(ch)
	}()
	for _, cli := range rpcs {
		go func(r *RpcClient) {
			if err := recover(); err != nil {
			}
			nonce, err := r.BaseFee()
			if err != nil {
				log.Info(err.Error())
				return
			}
			ch <- nonce
		}(cli)
	}
	select {
	case n = <-ch:
		return n, nil
	case <-time.After(time.Second * 10):
		return 0, errors.New("获取BaseFee超时")
	}
}
func GetBalance(rpcs []*RpcClient, addr string) (ret decimal.Decimal, err error) {
	ch := make(chan decimal.Decimal)

	defer func() {
		close(ch)
	}()
	for _, cli := range rpcs {
		go func(r *RpcClient) {
			if err := recover(); err != nil {
			}
			balance, err := r.GetBalance(addr)
			if err != nil {
				log.Info(err.Error())
				return
			}
			ch <- balance
		}(cli)
	}
	select {
	case ret = <-ch:
		return ret, nil
	case <-time.After(time.Second * 10):
		return ret, errors.New("获取balance超时")
	}
}

func SendRawtransaction(rpcs []*RpcClient, rawtx interface{}) (txid string, err error) {
	ch := make(chan string)
	defer func() {
		close(ch)
	}()
	for _, cli := range rpcs {
		go func(r *RpcClient) {
			if err := recover(); err != nil {
			}
			txhash, err := r.SendRawTransaction(rawtx)
			if err != nil {
				log.Info(err.Error())
				return
			}
			ch <- txhash
		}(cli)
	}
	select {
	case txid = <-ch:
		return txid, nil
	case <-time.After(time.Second * 10):
		return txid, errors.New("获取balance超时")
	}
}
