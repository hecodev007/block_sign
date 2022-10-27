package services

import (
	"dogeDataServer/common"
	"dogeDataServer/common/conf"
	"dogeDataServer/common/log"
	dao "dogeDataServer/models/po/doge"
	"dogeDataServer/utils"
	"fmt"
	"sync"
	"time"
)

type cacheTask struct {
	common.ProcTask
	Reconfirm bool
}
type ProcServer struct {
	processor   common.Processor
	IrserTasks  sync.Map
	StartHeight int64
	mempool     map[string]cacheTask
	mplock      sync.RWMutex
	stop        chan struct{}
	done        chan struct{}
}

func NewProcServer(processor common.Processor, bufNum int64) *ProcServer {
	startHeight := conf.Cfg.Sync.InitHeight
	dbHeight, _ := dao.MaxBlockHeight()
	if startHeight < dbHeight+1 {
		startHeight = dbHeight + 1
	}

	return &ProcServer{
		processor:   processor,
		mempool:     make(map[string]cacheTask),
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
		StartHeight: -1,
	}
}
func (s *ProcServer) mpload(k string) (cacheTask, bool) {
	s.mplock.RLock()
	defer s.mplock.RUnlock()
	ret, ok := s.mempool[k]
	return ret, ok
}
func (s *ProcServer) mpset(k string, v cacheTask) {
	s.mplock.Lock()
	defer s.mplock.Unlock()
	s.mempool[k] = v
}
func (s *ProcServer) mpdel(k string) {
	s.mplock.Lock()
	defer s.mplock.Unlock()
	delete(s.mempool, k)
}

func (s *ProcServer) Start() {
	if err := s.processor.Init(); err != nil {
		panic(fmt.Sprintf("initProcess, err :%v", err.Error()))
	}
	log.Infof("proc server start :%d", s.StartHeight)
	run := true
	go func() {
		for s.StartHeight < 0 {
			time.Sleep(time.Second * 1)
		}
		for run {
			select {
			case <-s.stop:
				log.Infof("proc server stop")
				run = false
				break
			default:
				//startime := time.Now()
				//加载task
				bullet := s.StartHeight / 100
				value, ok := s.IrserTasks.Load(bullet)
				if !ok {
					//log.Info("获取but失败等待", s.StartHeight, bullet)
					time.Sleep(time.Second * 3)
					break
				}
				tkmap := value.(*sync.Map)
				tkvalue, ok := tkmap.Load(s.StartHeight)
				if !ok {
					//log.Info("获取高度失败等待", s.StartHeight, bullet)
					time.Sleep(time.Second * 1)
					break
				}
				t := tkvalue.(common.ProcTask)
				////加载end
				//删除缓存
				if _, ok := s.mpload(t.GetBlockHash()); ok {
					s.mpdel(t.GetBlockHash())
				}
				if t.GetHeight()%10 == 0 {
					log.Info("处理不可逆交易", t.GetHeight(), t.GetBestHeight())
				}
				//处理不可逆交易
				if err := s.processor.ProcIrreverseTask(t); err != nil {
					log.Warn(err.Error())
				}

				s.StartHeight++
				//删除已经处理过的数据
				if s.StartHeight%100 == 0 {
					s.IrserTasks.Delete(bullet)
				}
				//log.Infof("processTask : %d , used time: %f 's", s.StartHeight, time.Since(startime).Seconds())
				break
			}
		}

		s.processor.Clear()
		log.Info("proc server shutdown")
		s.done <- struct{}{}
	}()
}
func (s *ProcServer) SetStartHeight(h int64) {
	s.StartHeight = h
}
func (s *ProcServer) Stop() {
	close(s.stop)
	<-s.done
}

//处理可逆任务
func (s *ProcServer) AddReverseTask(t common.ProcTask) {
	log.Info("处理可逆块：", t.GetHeight())

	//第一次推送交易
	if v, ok := s.mpload(t.GetBlockHash()); !ok {
		hasWatchAddr, err := s.processor.ProcReverseTxs(t)
		if err != nil {
			log.Error(err)
		}
		//log.Info(hasWatchAddr, t.GetBlockHash())
		s.mpset(t.GetBlockHash(), cacheTask{ProcTask: t, Reconfirm: hasWatchAddr})
		//后面推送确认数
	} else if v.Reconfirm && t.GetConfirms() > v.GetConfirms() {
		s.processor.PushReverseConfirms(t)
		s.mpset(t.GetBlockHash(), cacheTask{ProcTask: t, Reconfirm: true})
		return
	} else {
		return
	}
}

//先添加不可逆任务，异步处理
func (s *ProcServer) AddIrrevTask(t common.ProcTask) {
	height := t.GetHeight()
	but := height / 100

	if height%100 == 0 { //如果只根据hash删除,一些孤块删除不了
		s.mplock.Lock()
		for k, v := range s.mempool {
			if v.GetHeight() < height {
				delete(s.mempool, k)
			}
		}
		s.mplock.Unlock()
	}

	//防止数据过多，内存爆炸
	for _, ok := s.IrserTasks.Load(but - 2); ok; _, ok = s.IrserTasks.Load(but - 2) {
		time.Sleep(time.Second * 3)
	}
	utils.Lock("AddIrrevTask")
	defer utils.Unlock("AddIrrevTask")
	var tkmap = new(sync.Map)
	tk, ok := s.IrserTasks.Load(but)
	if ok {
		tkmap = tk.(*sync.Map)
	} else {
		s.IrserTasks.Store(but, tkmap)
	}

	tkmap.Store(height, t)

	//log.Info("AddIrrevTask success:", t.GetHeight())
}

func (s *ProcServer) AddProcTask(t common.ProcTask) {
}

func (s *ProcServer) SetPusher(pusher common.Pusher) *ProcServer {
	s.processor.SetPusher(pusher)
	return s
}

func (s *ProcServer) RemovePusher() {
	s.processor.RemovePusher()
}

func (s *ProcServer) RepushTx(userid int64, txid string) error {
	return s.processor.RepushTx(userid, txid)
}
