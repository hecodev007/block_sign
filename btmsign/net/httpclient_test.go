package net

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("before")
	m.Run()
}

type TestParams struct {
	Asset             string `json:"asset"`
	Transferreference string `json:"transferReference"`
}

func TestHttpClient(t *testing.T) {
	p := TestParams{Asset: "BTC", Transferreference: "d7ed366dad52a7dc9548da208ae40d5aee63e94bac66769309ebd1276f6272d0:3LZQ7K3CvTg718pHmLpCpiV8k7QzGCVRUa"}
	s := []TestParams{}
	s = append(s, p)
	c, _ := Post(TransferReceived("2e3"), s)
	fmt.Println(c)
}

func TestHttpClient2(t *testing.T) {
	p := TestParams{Asset: "ETH", Transferreference: "0x4b4bc77c42793642e1eb24f09078e55611a8f1ea7b8e19569d0e2e5f02954f8a:0x9696f59e4d72e237be84ffd425dcad154bf96976"}
	s := []TestParams{}
	s = append(s, p)
	c, _ := Post(TransferSent("2"), s)
	fmt.Println(c)
}

func TestHttpClient3(t *testing.T) {
	//p := TestParams{Asset: "BTC", Transferreference: "d7ed366dad52a7dc9548da208ae40d5aee63e94bac66769309ebd1276f6272d0:3LZQ7K3CvTg718pHmLpCpiV8k7QzGCVRUa"}
	//s := []TestParams{}
	//s = append(s, p)
	c, _ := Get(TransferSent("2"))
	fmt.Println(c)
}

func TestHttpClient4(t *testing.T) {
	//p := TestParams{Asset: "BTC", Transferreference: "d7ed366dad52a7dc9548da208ae40d5aee63e94bac66769309ebd1276f6272d0:3LZQ7K3CvTg718pHmLpCpiV8k7QzGCVRUa"}
	//s := []TestParams{}
	//s = append(s, p)
	c, _ := Post("http://13.231.191.20:19887/list-accounts", nil)
	fmt.Println(c)
}
