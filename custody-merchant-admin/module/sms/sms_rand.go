package sms

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func GetRands(do string) (string, string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	c := strconv.Itoa(r.Intn(899999) + 100000)
	m := fmt.Sprintf("【虎符】您的%s验证码是：%s，5分钟内有效，请勿泄露。", do, c)
	return m, c
}

func GetRandsAndEn() (string, string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	c := strconv.Itoa(r.Intn(899999) + 100000)
	m := fmt.Sprintf("[HOO] code: %s. Valid for 5 minutes.", c)
	return m, c
}

func GetEnRands(code string) string {
	msg := fmt.Sprintf("[HOO] code: %s. Valid for 5 minutes.", code)
	return msg
}
