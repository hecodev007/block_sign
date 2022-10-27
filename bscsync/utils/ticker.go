package utils

import "time"

type Fn func()

type MyTicker struct {
	MyTick *time.Ticker
	Runner Fn
}

func NewMyTicker(interval int, f Fn) *MyTicker {
	return &MyTicker{
		MyTick: time.NewTicker(time.Duration(interval) * time.Second),
		Runner: f,
	}
}

func (t *MyTicker) Start() {
	for {
		select {
		case <-t.MyTick.C:
			t.Runner()
		}
	}
}

func (t *MyTicker) Stop() {
	t.MyTick.Stop()
}
