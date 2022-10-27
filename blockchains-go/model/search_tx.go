package model

import (
	"errors"
	"time"
)

type SearchTx struct {
	Sign      string `json:"sign" form:"sign"`
	Sfrom     string `json:"sfrom" form:"sfrom"`
	Type      int    `json:"type" form:"type"`
	CoinName  string `json:"coin_name" form:"coin_name"`
	DateStart string `json:"date_start" form:"date_start"`
	DateEnd   string `json:"date_end" form:"date_end"`
}

func (s *SearchTx) Check() error {
	if s.Sfrom == "" {
		return errors.New("miss sfrom")
	}
	if s.CoinName == "" {
		return errors.New("miss coin_name")
	}

	//交易类型 可选  1 入账  2 出账  默认为1
	if s.Type != 1 || s.Type == 2 {
		if s.Type == 0 {
			s.Type = 1
		}
		return errors.New("error type")
	}
	//开始时间   默认一个月前
	if s.DateStart == "" {
		s.DateStart = time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	}
	//结束时间  默认当前时间
	if s.DateEnd == "" {
		s.DateEnd = time.Now().Format("2006-01-02 15:04:05")
	}
	return nil
}
