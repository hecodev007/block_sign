package services

import (
	"dataserver/common"
	"dataserver/log"

	//"dataserver/log"
	"dataserver/conf"
	"fmt"
	"strings"
	"time"
)

type ScanServer struct {
	scanner         common.Scanner
	processor       *ProcServer
	stop            chan struct{}
	done            chan struct{}
	conf            conf.SyncConfig
	irrevScanHeight int64
	nodeUrl         string
}

func NewScanServer(scanner common.Scanner, conf conf.Config, nodeUrl string) (*ScanServer, error) {
	return &ScanServer{
		scanner: scanner,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		conf:    conf.Sync,
		nodeUrl: nodeUrl,
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
					log.Infof("ScanData err: %v", err)
				}
				// time.Sleep(time.Second * 5)
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

	// 链高度
	bestBlockHeight, err := s.scanner.GetBestBlockHeight()
	if err != nil {
		log.Errorf("GetBlockCount %v", err)
		return err
	}

	// 2.获取db入账的区块高度
	dbBlockHeight, err := s.scanner.GetCurrentBlockHeight()
	if err != nil {
		log.Infof("%v", err)
	}
	log.Infof("bestBlockHeight : %d , dbBlockHeight : %d", bestBlockHeight, dbBlockHeight)

	if s.irrevScanHeight <= dbBlockHeight {
		s.irrevScanHeight = dbBlockHeight + 1
	}

	if s.irrevScanHeight < s.conf.InitHeight {
		s.irrevScanHeight = s.conf.InitHeight
	}

	gapCount := bestBlockHeight - s.irrevScanHeight
	if gapCount <= 0 {
		// 如果达到最高高度就睡眠10s时间
		time.Sleep(time.Second * time.Duration(s.conf.IntervalTime))
		return fmt.Errorf("already sacn reach best height: %d, current: %d", bestBlockHeight, s.irrevScanHeight)
	}

	// 如果扫描区块数量小于确认数，那么就开启内存池扫描
	if gapCount <= s.conf.Confirmations {
		// 已达到最高度，减缓扫描频率
		time.Sleep(time.Second * time.Duration(s.conf.IntervalTime))
		return nil
	}

	// 不可逆扫描的区块数量
	irreverseScanCount := gapCount - s.conf.Confirmations
	// 如果不可逆区块数量大于0，那么就开启不可逆扫描
	// 确定一个周期的轮询次数
	if irreverseScanCount > s.conf.EpochCount {
		irreverseScanCount = s.conf.EpochCount
	}

	// 如果一次扫描块数大于０，那么就开启多协程
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
					s.processor.AddProcTask(task.(common.ProcTask))
				}
			}
		}
	} else {
		for endHeight := s.irrevScanHeight + irreverseScanCount; s.irrevScanHeight < endHeight; s.irrevScanHeight++ {
			if strings.Contains(s.nodeUrl, "binance.org") {
				if s.irrevScanHeight&1 == s.conf.Flag {
					log.Infof("conf.Flag=%d, height=%d 执行该处理区块", s.conf.Flag, s.irrevScanHeight)
				} else {
					log.Infof("conf.Flag=%d, height=%d 跳过处理该处理区块", s.conf.Flag, s.irrevScanHeight)
					continue
				}
			}
			log.Infof("准备扫描区块 block height:%d", s.irrevScanHeight)
			task, err := s.scanner.ScanIrreverseBlock(s.irrevScanHeight, bestBlockHeight)
			if err != nil {
				log.Errorf("scanBlock :%s", err.Error())
				break
			}

			log.Infof("准备发送Task block height:%d count:%d", s.irrevScanHeight, len(task.GetTxs()))
			if s.processor != nil {
				s.processor.AddProcTask(task)
			}
			log.Infof("处理完毕Task block height:%d", s.irrevScanHeight)
			endtime = time.Now().Unix()
			if endtime-starttime > s.conf.EpochTime {
				break
			}
		}
	}

	endtime = time.Now().Unix()
	log.Infof("scanData ,block count:%d with use time : %d s ", irreverseScanCount, endtime-starttime)
	return nil
}
