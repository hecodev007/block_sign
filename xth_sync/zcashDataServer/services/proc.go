package services

import (
	"time"
	"zcashDataServer/common"
	"zcashDataServer/common/log"
)

type ProcServer struct {
	processor common.Processor
	procTasks chan common.ProcTask
	mempool   map[string]common.ProcTask
	stop      chan struct{}
	done      chan struct{}
}

func NewProcServer(processor common.Processor, bufNum int64) (*ProcServer, error) {
	return &ProcServer{
		processor: processor,
		procTasks: make(chan common.ProcTask, bufNum),
		mempool:   make(map[string]common.ProcTask),
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}, nil
}

func (s *ProcServer) Start() {
	if err := s.processor.Init(); err != nil {
		log.Errorf("initProcess, err :%v", err)
		return
	}
	log.Infof("proc server start ")

	run := true
	go func() {

		for run {
			select {
			case <-s.stop:
				log.Infof("proc server stop")
				run = false
				break
			case task := <-s.procTasks:

				starttime := time.Now()

				if task.GetIrreversible() {
					//检查数据库是否已经有记录该区块，有则直接退出
					if err := s.processor.CheckIrreverseBlock(task.GetBlockHash()); err != nil {
						continue
					}
					//查找内存池是否有该交易，如果无则推交易，如果有则推确认数
					_, ok := s.mempool[task.GetBlockHash()]
					if ok {
						// 更新确认数
						s.processor.UpdateReverseConfirms(task.GetBlock())
						//释放内存池区块
						delete(s.mempool, task.GetBlockHash())
					} else {
						//批量处理交易,这里其实应该做事务处理
						if err := s.processor.ProcIrreverseTxs(task.GetTxs(), task.GetBestHeight()); err != nil {
							//处理块
							//log.Infof("Proc irreverse Txs err : %v", err)
						}
						if err := s.processor.ProcIrreverseBlock(task.GetBlock()); err != nil {
							continue
						}
					}
					//更新确认数
					s.processor.UpdateIrreverseConfirms()
				} else {
					//查找内存池是否有该交易，如果无则推交易，如果有则推确认数
					oldTask, ok := s.mempool[task.GetBlockHash()]
					if !ok { //推送新交易
						if err := s.processor.ProcReverseTxs(task.GetTxs(), task.GetBestHeight()); err != nil {
							continue
						}
						s.mempool[task.GetBlockHash()] = task
					} else {
						//如果确认数更新了，通知消费方
						if task.GetConfirms() > oldTask.GetConfirms() {
							s.processor.UpdateReverseConfirms(task.GetBlock())
						}
					}
				}
				_ = starttime
				//log.Infof("processTask : %d ,tx: %d , used time: %f 's", task.GetHeight(), len(task.GetTxs()), time.Since(starttime).Seconds())
				break
			}
		}

		s.processor.Clear()
		log.Warn("proc server shutdown")
		s.done <- struct{}{}
	}()
}

func (s *ProcServer) Stop() {
	close(s.stop)
	<-s.done
}

func (s *ProcServer) AddProcTask(t common.ProcTask) {
	s.procTasks <- t
}

func (s *ProcServer) SetPusher(pusher common.Pusher) {
	s.processor.SetPusher(pusher)
}

func (s *ProcServer) RemovePusher() {
	s.processor.RemovePusher()
}

func (s *ProcServer) RepushTx(userid int64, txid string) error {
	return s.processor.RepushTx(userid, txid)
}
