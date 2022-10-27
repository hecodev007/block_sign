package util

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

type Status int

const (
	taskChanCap     = 10
	maxTaskCount    = 10
	maxIdleDuration = 10
	unKnowCatagory  = "unknow"
)
const (
	StatusInit Status = iota
	StatusReady
	StatusRun
	StatusHold
	StatusDone
	StatusFinish
)

type TaskState struct {
	*sync.Mutex
	TaskId          int64
	TaskCatagory    string
	TaskStatus      Status
	Counter         int64
	LastReleaseTime time.Time
	CreateTime      time.Time
	UpdateTime      time.Time
	FinishTime      time.Time
}

type Pool struct {
	Name string
	//WorkerFunc func(c net.Conn) error

	MaxTaskCount int

	MaxIdleDuration time.Duration

	idSeed int64

	Logger *log.Logger

	lock      *sync.Mutex
	taskCount int

	ready []*taskChan

	taskStates map[int64]*TaskState

	stopCh chan struct{}

	//chanPool        sync.Pool
}

type taskChan struct {
	status *TaskState
	ch     chan func()
}

func NewPool(name string, maxTaskCount int, maxIdleDuration int64, log *log.Logger) *Pool {
	return &Pool{
		Name:            name,
		MaxTaskCount:    maxTaskCount,
		MaxIdleDuration: time.Duration(maxIdleDuration),
		idSeed:          0,
		Logger:          log,
		lock:            &sync.Mutex{},
		taskCount:       0,
		ready:           []*taskChan{},
		taskStates:      map[int64]*TaskState{},
	}

}

func (p *Pool) Start() {
	if p.stopCh != nil {
		p.Logger.Errorln(p.Name + " Pool already started")
		return
	}

	if p.MaxTaskCount <= 0 || p.MaxTaskCount > 100 {
		p.MaxTaskCount = maxTaskCount
	}

	if p.MaxIdleDuration <= 0 {
		p.MaxIdleDuration = maxIdleDuration * time.Second
	}
	p.stopCh = make(chan struct{})
	stopCh := p.stopCh
	go func() {
		var scratch []*taskChan
		for {
			p.clean(&scratch)
			select {
			case <-stopCh:
				return
			default:
				time.Sleep(p.MaxIdleDuration)
			}
		}
	}()
}

func (p *Pool) Stop() {
	if p.stopCh == nil {
		p.Logger.Errorln(p.Name + " Pool wasn't started")
		return
	}
	close(p.stopCh)
	p.stopCh = nil

	// Stop all the workers waiting for incoming connections.
	// Do not wait for busy workers - they will stop after
	// serving the connection and noticing wp.mustStop = true.
	p.lock.Lock()
	ready := p.ready
	for i, ch := range ready {
		close(ch.ch)
		ch.ch = nil
		ready[i] = nil
	}
	p.ready = ready[:0]
	p.lock.Unlock()
}

//func (p *Pool) getMaxIdleDuration() time.Duration {
//	if p.MaxIdleDuration <= 0 {
//		return 10 * time.Second
//	}
//	return p.MaxIdleDuration
//}

func (p *Pool) clean(scratch *[]*taskChan) {
	maxIdleDuration := p.MaxIdleDuration

	// Clean least recently used workers if they didn't serve connections
	// for more than maxIdleWorkerDuration.
	currentTime := time.Now()

	p.lock.Lock()
	ready := p.ready
	n := len(ready)
	i := 0
	for i < n && currentTime.Sub(ready[i].status.LastReleaseTime) > maxIdleDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i = m; i < n; i++ {
			ready[i] = nil
		}
		p.ready = ready[:m]
	}
	p.lock.Unlock()

	// Notify obsolete workers to stop.
	// This notification must be outside the wp.lock, since ch.ch
	// may be blocking and may consume a lot of time if many workers
	// are located on non-local CPUs.
	// 改了,关闭这些通道释放通道
	tmp := *scratch
	for i, ch := range tmp {
		close(ch.ch)
		ch.ch = nil
		tmp[i] = nil
	}
}

func (p *Pool) RunTask(catagory string, f func()) (int64, bool) {
	ch := p.getCh()
	if ch == nil {
		return -1, false
	}
	ch.status.TaskCatagory = catagory
	ch.status.TaskStatus = StatusReady
	ch.ch <- f
	return ch.status.TaskId, true
}

func (p *Pool) getCh() *taskChan {
	var ch *taskChan
	createTask := false

	p.lock.Lock()
	ready := p.ready
	n := len(ready) - 1
	if n < 0 {
		if p.taskCount < p.MaxTaskCount {
			createTask = true
			p.taskCount++
		} else {
			p.Logger.Errorln(p.Name + " Pool max task count limit.")
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		p.ready = ready[:n]
	}
	p.lock.Unlock()

	if ch == nil {
		if !createTask {
			return nil
		}
		//新建线程channel
		ts := &TaskState{
			Mutex:        &sync.Mutex{},
			TaskId:       p.genTaskId(),
			TaskCatagory: unKnowCatagory,
			TaskStatus:   StatusInit,
			Counter:      0,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		ch = &taskChan{status: ts, ch: make(chan func(), taskChanCap)}
		p.taskStates[ts.TaskId] = ts
		go func() {
			p.runTask(ch)
		}()
	}
	return ch
}

//这里是释放入池,或是是删除
func (p *Pool) release(ch *taskChan) bool {
	ch.status.LastReleaseTime = time.Now()
	ch.status.TaskStatus = StatusDone
	ch.status.UpdateTime = ch.status.LastReleaseTime
	p.lock.Lock()
	p.ready = append(p.ready, ch)
	p.lock.Unlock()
	return true
}

func (p *Pool) runTask(ch *taskChan) {
	var f func()
	for f = range ch.ch {
		if f != nil {
			ch.status.Counter++
			ch.status.TaskStatus = StatusRun
			ch.status.UpdateTime = time.Now()
			//这里可以记录开始时间看运行
			f()
		}
		//这里可以记录结束时间看运行
		p.release(ch)
	}
	//如果线程不允许释放入池,则线程结束
	p.finish(ch)
}

func (p *Pool) finish(ch *taskChan) {
	ch.status.FinishTime = time.Now()
	ch.status.TaskStatus = StatusFinish
	//线程结束,删除在pool里登记的status
	delete(p.taskStates, ch.status.TaskId)
	p.lock.Lock()
	p.taskCount--
	p.lock.Unlock()
}

func (p *Pool) genTaskId() int64 {
	return atomic.AddInt64(&p.idSeed, 1)
}
