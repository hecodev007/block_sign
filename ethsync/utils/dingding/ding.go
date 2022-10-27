package dingding

import (
	"context"
	"encoding/json"
	"ethsync/common/log"
	"ethsync/utils/dingding/dingbot"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

type DingLevel int

const (
	Warn  DingLevel = 1
	Error DingLevel = 2
	Info  DingLevel = 3
)

type DingBot struct {
	Name   string
	Token  string
	Source chan []byte
}

var (
	ErrorDingBot *DingBot
)

// https://oapi.dingtalk.com/robot/send?access_token=2d16c6e586d84afdd422c131244bf35cd981a177bf3772a5be3b1436e1ed9526
func InitDingBot(ctx context.Context) {
	ErrorDingBot = &DingBot{
		Name:   "ding-robot-error",
		Token:  "5b77c51139f38a0b6bb98986a5dd1401c079efc11ec3d8f998c91ae18e290020",
		Source: make(chan []byte, 50),
	}
	go ErrorDingBot.Start(ctx)
}

func (s *DingBot) Start(ctx context.Context) {
loop:
	for {
		select {
		case msg := <-s.Source:
			resp, err := dingbot.SimpleTextMessage(string(msg)).Send(s.Token)
			if err != nil {
				log.Errorf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), err.Error())
			} else if resp.ErrCode != 0 && resp.ErrMsg != "" {
				dd, _ := json.Marshal(resp)
				log.Errorf("返回异常内容：%s", string(dd))
				switch resp.ErrCode {
				case 130101:
					// 20次限制，休息60秒
					log.Warn("send too fast,wait 60 seconds ")
					time.Sleep(60 * time.Second)
					// 继续发送一次
					resp, err = dingbot.SimpleTextMessage(string(msg)).Send(s.Token)
					if err != nil || resp.ErrCode != 0 || (resp.ErrCode == 0 && resp.ErrMsg == "") {
						// 抛弃发送
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}
						log.Errorf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), errMsg)
					}
				}
			} else {
				log.Infof("钉钉发送成功，发送内容：%s", string(msg))
			}
		case <-ctx.Done():
			log.Info(s.Name, " quit!")
			break loop
		}
	}
}

func NotifyErrorForTx(txId, from, contract string) {
	var sb strings.Builder
	sb.WriteString("ETH 数据服务发现无效交易: \n")
	sb.WriteString("txId: " + txId + "\n")
	sb.WriteString("from: " + from + "\n")
	sb.WriteString("contract: " + contract + "\n")
	ErrorDingBot.NotifyStr(sb.String())
}

func NotifyError(msg string) {
	ErrorDingBot.NotifyStr(msg)
}

func (s *DingBot) Notify(data []byte) {
	s.Source <- data
}

func (s *DingBot) NotifyStr(str string) {
	s.Source <- []byte(str)
}

func sha256(data ...[]byte) []byte {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}
