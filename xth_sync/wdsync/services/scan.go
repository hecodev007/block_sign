package services

import (
	"wdsync/common"
	"wdsync/common/conf"
	"wdsync/common/log"
	dao "wdsync/models/po/fil"
	"wdsync/utils"
	"time"
)

type ScanServer struct {
	scanner         common.Scanner
	processor       *ProcServer
	stop            chan struct{}
	done            chan struct{}
	conf            conf.SyncConfig
	irrevScanHeight int64
	bestHeight      int64
}

func NewScanServer(scanner common.Scanner, conf conf.Config) *ScanServer {
	return &ScanServer{
		scanner: scanner,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		conf:    conf.Sync,
	}
}

func (s *ScanServer) Start() {
	var err error
	if err = s.scanner.Init(); err != nil {
		log.Errorf("initScan, err :%v", err)
		return
	}
	log.Infof("scan server start")
	go func() {
		run := true
		for run {
			select {
			case <-s.stop:
				log.Infof("scan server stop")
				run = false
				break
			default:
				if err = s.scanData(); err != nil {
					log.Errorf("ScanData err: %v", err)
					break
				}
				//s.sleepOrStop(1)
				//time.Sleep(time.Second * 5)
				break
			}
		}
		s.scanner.Clear()
		log.Infof("scan server shutdown")
		s.done <- struct{}{}
	}()
}
func (s *ScanServer) sleepOrStop(d int64) {
	for i := int64(0); i < d; i++ {
		select {
		case <-s.stop:
			return
		default:
			time.Sleep(time.Second * 1)
		}
	}

}
func (s *ScanServer) Stop() {
	close(s.stop)
	<-s.done
}

func (s *ScanServer) Rollback(height int64) {
	//s.scanner.Rollback(height)
}

func (s *ScanServer) SetProcessor(p *ProcServer) *ScanServer {
	s.processor = p
	return s
}

func (s *ScanServer) RemoveProcessor() {
	s.processor = nil
}

//params:上次链区块高度
func (s *ScanServer) scanData() (err error) {
	startime := time.Now()
	//根据最高高度判断是否需要扫块
	//log.Info("scan", s.bestHeight)
	if bestHeight, err := s.scanner.GetBestBlockHeight(); err != nil {
		log.Infof("GetBlockCount %v", err)
		return nil
	} else if s.bestHeight == bestHeight { //区块高度没有增长，不处理
		time.Sleep(time.Second * 3)
		return nil
	} else {
		s.bestHeight = bestHeight
	}
	//起始高度
	if s.irrevScanHeight == 0 {
		//获取db入账的区块高度
		dbBlockHeight, err := s.scanner.GetCurrentBlockHeight()
		if err != nil {
			log.Errorf("GetCurrentBlockHeight %v", err)
			return err
		}
		//log.Info("dbBlockHeight", dbBlockHeight)
		if s.irrevScanHeight <= dbBlockHeight {
			s.irrevScanHeight = dbBlockHeight + 1
		}
		if s.irrevScanHeight < s.conf.InitHeight {
			s.irrevScanHeight = s.conf.InitHeight
		}
		s.processor.SetStartHeight(s.irrevScanHeight)
	}

	//s.bestHeight = ST

	//log.Infof("当前扫描区块:%d,最高区块 : %d ", s.irrevScanHeight, s.bestHeight)

	//需要执行的块个数
	gapCount := s.bestHeight - s.irrevScanHeight + 1
	if gapCount <= 0 {
		//如果达到最高高度就睡眠一定时间,是否是配置有问题
		log.Warnf("already sacn reach best height: %d, current: %d", s.bestHeight, s.irrevScanHeight)
		time.Sleep(time.Second * 3)
		return nil
	}

	//headHeight, err := s.scanner.ChainHeadHeight()
	log.Infof("当前扫描区块:%d,Confirmations:%d,最高区块 : %d ", s.irrevScanHeight, s.conf.Confirmations, s.bestHeight)

	endHeight := s.bestHeight - s.conf.Confirmations
	if endHeight >= s.bestHeight {
		endHeight = s.bestHeight - 1
	}
	//先top=>parent=>parent遍历头
	if dao.CountBlock(s.irrevScanHeight, s.bestHeight) != s.bestHeight-s.irrevScanHeight+1 {
		//log.Info(s.irrevScanHeight, s.bestHeight, dao.CountBlock(s.irrevScanHeight, s.bestHeight))
		if err = s.scanner.ScanHead(s.irrevScanHeight); err != nil {
			log.Info(err.Error())
			return err
		}
	}

	workPool2 := utils.NewWorkPool(5)
	for ; s.irrevScanHeight < endHeight; s.irrevScanHeight++ {
		workPool2.Incr()
		go func(irrevScanHeight, bestHeight int64) {
			defer workPool2.Done()
			//log.Info(irrevScanHeight)
			//defer log.Info(irrevScanHeight)
			task, err := s.scanner.ScanBaseBlock(irrevScanHeight, bestHeight)
			if err != nil {
				log.Info(err.Error())
				return
			}

			if s.processor != nil {
				s.processor.AddIrrevTask(task)
			}
		}(s.irrevScanHeight, s.bestHeight)

	}
	workPool2.Wait()
	//处理时间超过一分钟不扫可逆块
	if time.Since(startime) > time.Minute {
		return nil
	}
	if endHeight < s.irrevScanHeight {
		endHeight = s.irrevScanHeight
	}
	//开启可逆扫块
	//log.Info(conf.Cfg.Sync.EnableMempool)
	if conf.Cfg.Sync.EnableMempool {
		for ; endHeight < s.bestHeight; endHeight++ {
			workPool2.Incr()
			go func(endHeight, bestHeight int64) {
				defer workPool2.Done()
				task, err := s.scanner.ScanBaseBlock(endHeight, bestHeight)
				if err != nil {
					return
				}
				if s.processor != nil {
					s.processor.AddReverseTask(task)
				}
			}(endHeight, s.bestHeight)
		}
		workPool2.Wait()
	}
	//time.Sleep(time.Second * 2)
	return nil
}

