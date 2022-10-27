package dingding

import (
	"brisesign/conf"
	"brisesign/redis"
	"brisesign/util/dingding/dingbot"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
	"strings"
	"time"
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
	WarnDingBot  *DingBot
	InfoDingBot  *DingBot
)

// https://oapi.dingtalk.com/robot/send?access_token=2d16c6e586d84afdd422c131244bf35cd981a177bf3772a5be3b1436e1ed9526
func InitDingBot(ctx context.Context) {
	if conf.Config.IMBot.DingErrorToken != "" {
		ErrorDingBot = &DingBot{
			Name:   "ding-robot-error",
			Token:  conf.Config.IMBot.DingErrorToken,
			Source: make(chan []byte, 50),
		}
		go ErrorDingBot.Start(ctx)
	} else {
		log.Println("不启用钉钉ERROR机器人：没有配置dingErrorToken")
	}

	if conf.Config.IMBot.DingWarnToken != "" {
		WarnDingBot = &DingBot{
			Name:   "ding-robot-warn",
			Token:  conf.Config.IMBot.DingWarnToken,
			Source: make(chan []byte, 50),
		}
		go WarnDingBot.Start(ctx)
	} else {
		log.Println("不启用钉钉WARN机器人：没有配置dingWarnToken")
	}

	if conf.Config.IMBot.DingInfoToken != "" {
		InfoDingBot = &DingBot{
			Name:   "ding-robot-info",
			Token:  conf.Config.IMBot.DingInfoToken,
			Source: make(chan []byte, 50),
		}
		go InfoDingBot.Start(ctx)
	} else {
		log.Println("不启用钉钉INFO机器人：没有配置dingInfoToken")
	}
}

func (s *DingBot) Start(ctx context.Context) {
loop:
	for {
		select {
		case msg := <-s.Source:
			resp, err := dingbot.SimpleTextMessage(string(msg)).Send(s.Token)
			if err != nil {
				log.Printf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), err.Error())
			} else if resp.ErrCode != 0 && resp.ErrMsg != "" {
				dd, _ := json.Marshal(resp)
				log.Printf("返回异常内容：%s", string(dd))
				switch resp.ErrCode {
				case 130101:
					// 20次限制，休息60秒
					log.Println("send too fast,wait 60 seconds ")
					time.Sleep(60 * time.Second)
					// 继续发送一次
					resp, err = dingbot.SimpleTextMessage(string(msg)).Send(s.Token)
					if err != nil || resp.ErrCode != 0 || (resp.ErrCode == 0 && resp.ErrMsg == "") {
						// 抛弃发送
						errMsg := ""
						if err != nil {
							errMsg = err.Error()
						}
						log.Printf("钉钉发送内容异常, 机器人：%s,发送内容：%s，返回错误:%s ", s.Name, string(msg), errMsg)
					}
				}
			} else {
				log.Printf("钉钉发送成功，发送内容：%s", string(msg))
			}
		case <-ctx.Done():
			log.Println(s.Name, " quit!")
			break loop
		}
	}
}

func NotifyWarn(title string, outerOrderNo string, sections ...string) {
	talk(Warn, title, outerOrderNo, sections...)
}

func NotifyError(title string, outerOrderNo string, sections ...string) {
	talk(Error, title, outerOrderNo, sections...)
}

func NotifyInfo(title string, outerOrderNo string, sections ...string) {
	talk(Info, title, outerOrderNo, sections...)
}

// 目前发送`warn`和`error`等级的通知
// 由于服务存在循环执行的任务，可能会导致一些完全相同的信息会频繁的重复通知
// 所以会以订单编号来作为redis的key，有效时间10分钟
// 对于没有订单编号的情况，取 SHA-3 sha256(msg)前24个字节作为key
func talk(level DingLevel, tag string, outerOrderNo string, sections ...string) {
	if Error == level && ErrorDingBot == nil {
		log.Printf("未初始化的机器人 %d", Error)
		return
	}
	if Warn == level && WarnDingBot == nil {
		log.Printf("未初始化的机器人 %d", Warn)
		return
	}
	if Info == level && InfoDingBot == nil {
		log.Printf("未初始化的机器人 %d", Info)
		return
	}

	var msg strings.Builder
	msg.WriteString(tag)
	msg.WriteString("\n")
	if outerOrderNo != "" {
		msg.WriteString("订单号: ")
		msg.WriteString(outerOrderNo)
		msg.WriteString("\n")
	}
	for _, section := range sections {
		if section == "" {
			continue
		}
		msg.WriteString("- ")
		msg.WriteString(section)
		msg.WriteString("\n")
	}

	key := ""
	if outerOrderNo == "" {
		key = hex.EncodeToString(sha256([]byte(msg.String())))[:24]
	} else {
		key = outerOrderNo
	}

	// 同一订单按不同级别区分
	interceptKey := fmt.Sprintf("bsc_sign_notify_%d_order_no_%s", level, key)
	setSuccess, err := redis.Client.SetNX(interceptKey, "", 600)
	if err != nil {
		log.Printf("notifyDingDing设置redis失败：%s", err.Error())
		return
	}
	if !setSuccess {
		return
	}

	log.Printf("ding talk setSuccess:%s", setSuccess)
	if setSuccess {
		switch level {
		case Warn:
			WarnDingBot.NotifyStr(msg.String())
		case Error:
			ErrorDingBot.NotifyStr(msg.String())
		case Info:
			InfoDingBot.NotifyStr(msg.String())
		default:
			WarnDingBot.NotifyStr(msg.String())
		}
		log.Println(msg.String())
	}
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
