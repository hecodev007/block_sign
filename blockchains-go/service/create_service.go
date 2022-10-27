package service

import (
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/prometheus/common/log"
	"sync"
	"time"
)

type Creator interface {
	Name() string
	CreateTx(entity *entity.FcTransfersApply) (int64, error)
}

type CreateService struct {
	creators      map[string]Creator
	createChannel map[string]chan *entity.FcTransfersApply
	stopCh        chan struct{}
	stopWg        *sync.WaitGroup
}

func NewCreateService(cs []Creator) *CreateService {
	svr := &CreateService{
		creators:      make(map[string]Creator),
		createChannel: make(map[string]chan *entity.FcTransfersApply),
		stopCh:        make(chan struct{}),
		stopWg:        &sync.WaitGroup{},
	}

	for _, v := range cs {
		svr.creators[v.Name()] = v
		svr.createChannel[v.Name()] = make(chan *entity.FcTransfersApply)
	}

	return svr
}

func (s *CreateService) Start() error {

	s.stopWg.Add(len(s.createChannel))
	//开启各个币种create的协程
	for k, v := range s.createChannel {
		go func(chan *entity.FcTransfersApply) {
			for {
				select {
				case <-s.stopCh:
					s.stopWg.Done()
					log.Warnln("autoCreateTx finish ")
					return
				case d := <-v:
					if c, ok := s.creators[k]; ok {
						if _, err := c.CreateTx(d); err != nil {
							log.Warnf("create tx err : %v", err)
						}
					} else {
						log.Warnf("don't support %s create", c)
					}
				}
			}
		}(v)
	}

	//开启订单的扫描
	for {
		select {
		case <-s.stopCh:
			log.Warnln("autoCreateTx finish ")
			break
		default:
			s.autoCreateTx()
			time.Sleep(time.Second * 2)
		}
	}

	return nil
}

func (s *CreateService) Stop() error {
	close(s.stopCh)
	s.stopWg.Wait()
	return nil
}

//数据库轮询订单任务
func (s *CreateService) autoCreateTx() {
	agreeApplys, err := dao.FcTransfersApplyFindByAgree()
	if err != nil {
		log.Warnf("find agree apply, err : %v", err)
		return
	}

	for _, v := range agreeApplys {

		err = dao.FcTransfersApplyUpdateStatusById(v.Id, 7)
		if err != nil {
			log.Warnf("update agree apply err : %v", err)
			continue
		}

		if dao.FcOrderHaveByOutOrderNo(v.OutOrderid, 4) {
			log.Warnf("already have fc order")
			continue
		}

		if dao.FcOrderHotHaveByOutOrderNo(v.OutOrderid, 4) {
			log.Warnf("already have fc order hot")
			continue
		}

		if c, ok := s.createChannel[v.CoinName]; ok {
			c <- v
		} else {
			log.Warnf("create err : %v", err)
		}
	}
}
