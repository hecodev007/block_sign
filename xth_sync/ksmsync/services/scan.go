package services

import (
	"encoding/json"
	"ksmsync/common"
	"ksmsync/common/conf"
	"ksmsync/common/log"
	"ksmsync/utils"
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
	s.scanner.Rollback(height)
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
	if bestHeight, err := s.scanner.GetBestBlockHeight(); err != nil {
		log.Errorf("GetBlockCount %v", err)
		time.Sleep(time.Second * 3)
		return nil
	} else if s.bestHeight == bestHeight { //区块高度没有增长，不处理
		//time.Sleep(time.Second * 3)
		return nil
	} else {
		s.bestHeight = bestHeight
	}

	if s.irrevScanHeight == 0 {
		//获取db入账的区块高度
		dbBlockHeight, err := s.scanner.GetCurrentBlockHeight()
		if err != nil {
			log.Errorf("GetCurrentBlockHeight %v", err)
			return err
		}

		if s.irrevScanHeight <= dbBlockHeight {
			s.irrevScanHeight = dbBlockHeight + 1
		}
		if s.irrevScanHeight < s.conf.InitHeight {
			s.irrevScanHeight = s.conf.InitHeight
		}

		s.processor.SetStartHeight(s.irrevScanHeight)
	}
	log.Infof("当前扫描区块:%d,最高区块 : %d ", s.irrevScanHeight, s.bestHeight)

	//需要执行的块个数
	gapCount := s.bestHeight - s.irrevScanHeight + 1
	if gapCount <= 0 {
		//如果达到最高高度就睡眠一定时间,是否是配置有问题
		log.Warnf("already sacn reach best height: %d, current: %d", s.bestHeight, s.irrevScanHeight)
		//time.Sleep(time.Second * (3 * time.Duration(int64((gapCount)*-1))))
		return nil
	}

	//log.Info("gapCount", gapCount, "s.conf.Confirmations", s.conf.Confirmations)
	//扫描不可逆区块
	if gapCount > s.conf.Confirmations {
		if s.conf.MultiScanNum <= 0 {
			s.conf.MultiScanNum = 10
		}

		workPool := utils.NewWorkPool(int(s.conf.MultiScanNum))
		for endHeight := s.bestHeight - s.conf.Confirmations; s.irrevScanHeight <= endHeight; s.irrevScanHeight++ {
			workPool.Incr()
			//log.Info("workPool start", workPool.Running, s.irrevScanHeight, endHeight)
			go func(height, bestHeight int64) {
				defer workPool.Dec()
			ScanIrreverseBlock:
				task, err := s.scanner.ScanIrreverseBlock(height, bestHeight)
				//log.Info("workPool end", workPool.Running, height, bestHeight)
				if err != nil {
					log.Info(height, err.Error())
					time.Sleep(time.Second*10)
					goto ScanIrreverseBlock
				}
				if s.processor != nil {
					//_ = task
					s.processor.AddIrrevTask(task)
				}
				//log.Info(height, bestHeight,String(task))
				time.Sleep(time.Millisecond*100)
			}(s.irrevScanHeight, s.bestHeight)

		}
		workPool.Wait()
	}
	//在处理不可逆区块期间交区块又有所增长
	if gapCount > s.conf.Confirmations+2 {
		if dbBlockHeight, _ := s.scanner.GetCurrentBlockHeight(); dbBlockHeight != s.bestHeight {
			return nil
		}
	}
	//如果激活内存池扫描，开启内存池扫描
	if s.conf.EnableMempool {
		//循环扫描所有已出但可逆的区块
		for height := s.irrevScanHeight; height <= s.bestHeight; height++ {
			tmp, err := s.scanner.ScanReverseBlock(height, s.bestHeight)
			if err != nil {
				log.Warnf("scanBlock err: %v height:%d bestheight:%d", err, height, s.bestHeight)
				continue
			}
			//log.Info(s.irrevScanHeight, s.bestHeight, tmp.GetIrreversible(), tmp.GetConfirms(), tmp.GetIrreversible())

			if s.processor != nil && !tmp.GetIrreversible() {
				//处理可逆任务
				s.processor.AddReverseTask(tmp)
			}
		}
	}
	log.Infof("scanData:%d's block, with use time : %d s \n\n", gapCount, int64(time.Since(startime).Seconds()))
	//已达到最高度，减缓扫描频率
	//time.Sleep(time.Second * 2)
	return nil

}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
