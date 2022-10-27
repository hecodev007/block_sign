package util

import (
	"fmt"
	"sync"
	"time"
)

const (
	nano         = 1000000
	workerIdBits = 8    // workerid 二进制位数(8位即255)
	maxWorkerId  = 255  // 最大的worker id
	sequenceBits = 12   // sequence 二进制位数(12位即4095)
	maxSequence  = 4095 // 最大的sequence号
)

var (
	ysSince int64 = time.Date(2013, 1, 0, 0, 0, 0, 0, time.UTC).UnixNano() / nano //英盛网成立时间
	since   int64 = 0
)

//声明id生成器变量时请用该接口类型解偶
type IdGenerator interface {
	Next() (uint64, error)
}

//workerId当前的worker的id号
//onlyDuration生成的id是否有完整的时间戳信息, true:从yssince算起的时间duration, false:完整
func NewSnowFlake(workerId uint32, onlyDuration bool) *SnowFlake {
	if workerId < 0 {
		workerId = 1
	} else if workerId > maxWorkerId {
		workerId = maxWorkerId
	}
	if onlyDuration {
		since = ysSince
	}
	return &SnowFlake{workerId: workerId}
}

type SnowFlake struct {
	lastTimestamp uint64
	workerId      uint32
	sequence      uint32
	lock          sync.Mutex
}

func (sf *SnowFlake) uint64() uint64 {
	return (sf.lastTimestamp << (workerIdBits + sequenceBits)) | (uint64(sf.workerId) << sequenceBits) | (uint64(sf.sequence))
}

func (sf *SnowFlake) Next() (uint64, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	ts := timestamp()
	if ts == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & maxSequence
		if sf.sequence == 0 {
			ts = tilNextMillis(ts)
		}
	} else {
		sf.sequence = 0
	}

	if ts < sf.lastTimestamp {
		return 0, fmt.Errorf("Invalid timestamp: %v - precedes %v", ts, sf)
	}

	sf.lastTimestamp = ts
	id := sf.uint64()
	return id, nil
}

func timestamp() uint64 {
	return uint64(time.Now().UnixNano()/nano - since)
}

func tilNextMillis(ts uint64) uint64 {
	i := timestamp()
	for i < ts {
		i = timestamp()
	}
	return i
}

//尽可能不要用这个生成work_id,用配置设置work_id------------
//func Default() (*SnowFlake, error) {
//	return NewSnowFlake(DefaultWorkId())
//}
//
//func DefaultWorkId() uint32 {
//	var id uint32
//	ift, err := net.Interfaces()
//	if err != nil {
//		rand.Seed(time.Now().UnixNano())
//		id = rand.Uint32() % MaxWorkerId
//	} else {
//		h := crc32.NewIEEE()
//		for _, value := range ift {
//			h.Write(value.HardwareAddr)
//		}
//		id = h.Sum32() % MaxWorkerId
//	}
//	return id & MaxWorkerId
//}
//------------------------------------------------------
