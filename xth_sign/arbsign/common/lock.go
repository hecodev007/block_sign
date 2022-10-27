package common

import (
	"errors"
	"sync"
)

var lockMap sync.Map

func init() {
	//lockMap = make(map[string]*sync.Mutex)
}

func Lock(k string) error {
	if v, ok := lockMap.Load(k); !ok {
		mtx := &sync.Mutex{}
		mtx.Lock()
		lockMap.Store(k, mtx)
	} else if l, ok := v.(*sync.Mutex); ok {
		l.Lock()
	} else {
		return errors.New("panic lock")
	}
	return nil
}

func Unlock(k string) error {
	if v, ok := lockMap.Load(k); !ok {
		return errors.New("panic unlock")
	} else if l, ok := v.(*sync.Mutex); !ok {
		return errors.New("panic unlock")
	} else {
		l.Unlock()
	}
	return nil
}
