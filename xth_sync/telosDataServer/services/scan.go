package services

import (
	"telosDataServer/common"
	"telosDataServer/common/log"
	"telosDataServer/conf"
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
			//log.Infof("scan server running")
			select {
			case <-s.stop:
				log.Infof("scan server stop")
				run = false
				break
			default:
				if err := s.scanData(); err != nil {
					log.Infof("ScanData err: %v", err)
				}
				time.Sleep(time.Second * 1)
				break
			}
		}
		s.scanner.Clear()
		log.Infof("scan server shutdown")
		s.done <- struct{}{}
	}()
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
	starttime := time.Now().Unix()
	endtime := starttime

	bestBlockHeight, err := s.scanner.GetBestBlockHeight()
	if err != nil {
		log.Errorf("GetBlockCount %v", err)
		return err
	}

	//2.获取db入账的区块高度
	dbBlockHeight, err := s.scanner.GetCurrentBlockHeight()
	if err != nil {
		log.Warnf("%v", err)
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
		time.Sleep(time.Second)
		return nil
		//return fmt.Errorf("already sacn reach best height: %d, current: %d", bestBlockHeight, s.irrevScanHeight)
	}
	//if bestBlockHeight%10==0{
		log.Infof("bestBlockHeight : %d , dbBlockHeight : %d,irrevScanHeight:%v", bestBlockHeight, dbBlockHeight,s.irrevScanHeight)
	//}
	//如果扫描区块数量小于确认数，那么就开启内存池扫描
	if gapCount <= s.conf.Confirmations {
		//如果激活内存池扫描，开启内存池扫描
		if s.conf.EnableMempool {
			//循环扫描所有已出但可逆的区块
			for i := int64(0); i < gapCount; i++ {
				height := s.irrevScanHeight + i
				tmp, err := s.scanner.ScanReverseBlock(height, bestBlockHeight)
				if err != nil {
					log.Warnf("scanBlock err: %v", err)
					break
				}

				if s.processor != nil {
					log.Infof("加入3")
					s.processor.AddProcTask(tmp)
				}
			}
		}
		//已达到最高度，减缓扫描频率
		time.Sleep(time.Second * time.Duration(s.conf.IntervalTime))
		return nil
	}

	//不可逆扫描的区块数量=
	irreverseScanCount := gapCount - s.conf.Confirmations
	//如果不可逆区块数量大于0，那么就开启不可逆扫描
	//确定一个周期的轮询次数
	if irreverseScanCount > s.conf.EpochCount {
		irreverseScanCount = s.conf.EpochCount
	}

	//如果一次扫描块数大于０，那么就开启多协程
	if s.conf.MultiScanNum > 0 {
		for endHeight := s.irrevScanHeight + irreverseScanCount; s.irrevScanHeight < endHeight; s.irrevScanHeight += s.conf.MultiScanNum {
			step := s.conf.MultiScanNum
			if step > irreverseScanCount {
				step = irreverseScanCount
			}
			taskmap := s.scanner.BatchScanIrreverseBlocks(s.irrevScanHeight, s.irrevScanHeight+step, bestBlockHeight)
			for i := int64(0); i < step; i++ {
				height := s.irrevScanHeight + i
				task, ok := taskmap.Load(height)
				if ok && s.processor != nil {
					log.Infof("加入2")
					s.processor.AddProcTask(task.(common.ProcTask))
				}
			}
		}
	} else {
		for endHeight := s.irrevScanHeight + irreverseScanCount; s.irrevScanHeight < endHeight; s.irrevScanHeight++ {
			//log.Infof("扫描高度：%d/%d", s.irrevScanHeight,endHeight)
			task, err := s.scanner.ScanIrreverseBlock(s.irrevScanHeight, bestBlockHeight)
			if err != nil {
				log.Errorf("ScanIrreverseBlock :%s", err.Error())
				break
			}
			//dd, err := json.Marshal(task.GetTxs())
			//if err != nil {
			//	log.Errorf("error :%s", err.Error())
			//}

			if s.processor != nil {
				//log.Infof("加入1")
				s.processor.AddProcTask(task)
			}

			endtime = time.Now().Unix()
			if endtime-starttime > s.conf.EpochTime {
				break
			}
		}
	}

	endtime = time.Now().Unix()
	//log.Infof("scanData , with use time : %d s ", endtime-starttime)
	return nil
}
