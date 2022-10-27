package dingding

import (
	"encoding/json"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/dingbot"
	"time"
)

type DingBot struct {
	Name   string
	Token  string
	Source chan []byte
	Quit   chan struct{}
}

var (
	ErrTransferDingBot *DingBot // 全局使用,报警机器人
	ReviewDingBot      *DingBot // 全局使用,审核机器人
	WarnDingBot        *DingBot
)

// https://oapi.dingtalk.com/robot/send?access_token=2d16c6e586d84afdd422c131244bf35cd981a177bf3772a5be3b1436e1ed9526
func InitDingBot() {
	if conf.Cfg.IMBot.SecretToken == "" {
		panic("miss DingSecretToken")
	}
	if conf.Cfg.IMBot.DingToken == "" {
		panic("miss DingToken")
	}
	if ErrTransferDingBot == nil {
		ErrTransferDingBot = &DingBot{
			Name:   "ding-robot-callback",
			Token:  conf.Cfg.IMBot.DingToken,
			Source: make(chan []byte, 50),
			Quit:   make(chan struct{}),
		}
		ErrTransferDingBot.Start()
	}

	if conf.Cfg.IMBot.ReviewToken == "" {
		panic("miss ReviewToken")
	}
	if ReviewDingBot == nil {
		ReviewDingBot = &DingBot{
			Name:   "ding-robot-review",
			Token:  conf.Cfg.IMBot.ReviewToken,
			Source: make(chan []byte, 50),
			Quit:   make(chan struct{}),
		}
		ReviewDingBot.Start()
	}

	if WarnDingBot == nil {
		WarnDingBot = &DingBot{
			Name:   "ding-robot-warn",
			Token:  conf.Cfg.IMBot.ReviewToken,
			Source: make(chan []byte, 50),
			Quit:   make(chan struct{}),
		}
		WarnDingBot.Start()
	}
}

func (s *DingBot) Start() {
	// w.Source = make(chan []byte, 50)
	go func() {
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
							log.Errorf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), err.Error())
						}
						// default:
						//	log.Errorf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), err.Error())
					}
				} else {
					log.Infof("钉钉发送成功，发送内容：%s", string(msg))
				}
			case <-s.Quit:
				log.Info(s.Name, " quit!")
				return
			}
		}
	}()
}

func (s *DingBot) stop() {
	s.Quit <- struct{}{}
}

func (s *DingBot) Notify(data []byte) {
	s.Source <- data
}

func (s *DingBot) NotifyStr(str string) {
	s.Source <- []byte(str)
}
