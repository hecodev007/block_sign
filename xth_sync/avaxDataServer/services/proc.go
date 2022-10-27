package services

import (
	"avaxDataServer/common"
	"avaxDataServer/common/log"
	"time"
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
					log.Info("不可逆区块:", task.GetHeight())
					//检查数据库是否已经有记录该区块，有则直接退出
					//if err := s.processor.CheckIrreverseBlock(task.GetBlockHash()); err != nil {
					//	log.Error(err.Error())
					//	continue
					//}

					//批量处理交易,这里其实应该做事务处理
					if err := s.processor.ProcIrreverseTxs(task); err != nil {
						log.Error(err.Error())
					}

					//更新确认数
					s.processor.UpdateIrreverseConfirms()
				} else {

				}
				log.Infof("processTask : %d ,tx: %d , used time: %f 's", task.GetHeight(), len(task.GetTxs()), time.Since(starttime).Seconds())
				break
			}
		}

		s.processor.Clear()
		log.Info("proc server shutdown")
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
