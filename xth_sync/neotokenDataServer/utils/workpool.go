package utils

import (
	"sync"
	"sync/atomic"
)

type WorkPool struct {
	Cond    *sync.Cond
	Wg      *sync.WaitGroup
	Max     int32 //最大goroutine个数
	Running int32 //正在运行goroutine个数
}

func NewWorkPool(max int) *WorkPool {
	L := new(sync.Mutex)
	return &WorkPool{
		Wg:      &sync.WaitGroup{},
		Cond:    sync.NewCond(L),
		Max:     int32(max),
		Running: 0,
	}
}

//设置最大goroutine
func (gp *WorkPool) Set(max int) {
	gp.Max = int32(max)
	gp.Cond.Signal()
}

//新增一个goroutine
func (gp *WorkPool) Incr() {
	gp.Add(1)
}
func (gp *WorkPool) Add(n int32) {
	gp.Cond.L.Lock()
	defer gp.Cond.L.Unlock()

	for gp.Running >= gp.Max {
		gp.Cond.Wait()
	}
	atomic.AddInt32(&gp.Running, n)
	gp.Wg.Add(int(n))
}
func (gp *WorkPool) Done() {
	gp.Dec()
}

//结束一个goroutine
func (gp *WorkPool) Dec() {
	atomic.AddInt32(&gp.Running, -1)
	gp.Wg.Done()
	gp.Cond.Signal()
}

//等待所有执行完毕
func (gp *WorkPool) Wait() {
	gp.Wg.Wait()
}
