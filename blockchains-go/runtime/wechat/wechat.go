package wechat

import (
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
)

func SendWarnInfo(content string) error {
	if content == "" {
		return fmt.Errorf("")
	}
	mapData := make(map[string]string)
	mapData["content"] = content
	result, err := util.PostJson(conf.Cfg.WeChat.Url, mapData)
	log.Debugf("返回结果：%s", string(result))
	if err != nil {
		return err
	}
	return nil
}
