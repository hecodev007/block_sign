package services

import (
	"avaxDataServer/common"
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	"time"
)

type ScanServer struct {
	scanner         common.Scanner
	processor       *ProcServer
	stop            chan struct{}
	done            chan struct{}
	conf            conf.SyncConfig
	irrevScanHeight int64
}

func NewScanServer(scanner common.Scanner, conf conf.Config) (*ScanServer, error) {
	return &ScanServer{
		scanner: scanner,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		conf:    conf.Sync,
	}, nil
}

func (s *ScanServer) Start() {
	if err := s.scanner.Init(); err != nil {
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
				if err := s.scanData(); err != nil {
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

func (s *ScanServer) SetProcessor(p *ProcServer) {
	s.processor = p
}

func (s *ScanServer) RemoveProcessor() {
	s.processor = nil
}

func (s *ScanServer) scanData() error {
	//:= time.Now()

	bestBlockHeight, err := s.scanner.GetBestBlockHeight()
	if err != nil {
		log.Errorf("GetBlockCount %v", err)
		return err
	}

	//2.获取db入账的区块高度
	dbBlockHeight, err := s.scanner.GetCurrentBlockHeight()
	if err != nil {
		log.Error("%v", err)
		return err
	}

	if s.irrevScanHeight <= dbBlockHeight {
		s.irrevScanHeight = dbBlockHeight + 1
	}

	if s.irrevScanHeight < s.conf.InitHeight {
		s.irrevScanHeight = s.conf.InitHeight
	}

	gapCount := bestBlockHeight - s.irrevScanHeight
	if gapCount <= 0 {
		//如果达到最高高度就睡眠10s时间
		time.Sleep(time.Second * 10)
		return nil
		//return fmt.Errorf("already sacn reach best height: %d, current: %d", bestBlockHeight, s.irrevScanHeight)
	}
	log.Infof("当前执行区块:%d,最高区块 : %d , 数据库区块 : %d", s.irrevScanHeight, bestBlockHeight, dbBlockHeight)

	//log.Info("gapCount", gapCount, "s.conf.Confirmations", s.conf.Confirmations)
	//如果扫描区块数量小于确认数，那么就开启内存池扫描

	//log.Infof("扫描高度：%d ", s.irrevScanHeight)
	task, err := s.scanner.ScanIrreverseBlock(s.irrevScanHeight, bestBlockHeight)
	if err != nil {
		log.Errorf("error :%s", err.Error())
	}
	if task == nil {
		panic("")
	}
	if s.processor != nil {
		//log.Infof("加入1 confirms:%d", task.(common.ProcTask).GetConfirms())
		s.processor.AddProcTask(task)
	}

	//time.Sleep(time.Second*200000)
	time.Sleep(time.Second * 2)
	//log.Infof("scanData , with use time : %d s ", int64(time.Since(starttime).Seconds()))
	return nil
}
