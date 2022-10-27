package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bsc-sign/conf"
	"github.com/bsc-sign/model"
	"github.com/bsc-sign/redis"
	"github.com/bsc-sign/util/log"
	"time"
)

const (
	// 等待执行的订单列表 缓存key
	orderKeeperListKey = "bsc_sign_keeper_list"

	// 当前已接收处理的订单 缓存key
	// 用于检查订单是否重复请求
	orderKeeperAgainstReplay = "bsc_sign_against_replay_"
)

type OrderKeeper struct {
	ctx context.Context

	// orderKeeperAccept缓存的有效期限
	againstReplayCacheExpiration time.Duration

	// 处理订单的最大数量，正在处理 + 待处理
	keeperSize int64
}

func NewOrderKeeper(ctx context.Context) *OrderKeeper {
	return &OrderKeeper{
		ctx:                          ctx,
		againstReplayCacheExpiration: time.Duration(conf.Config.OrderKeeper.CacheExpirationSec) * time.Second,
		keeperSize:                   conf.Config.OrderKeeper.KeeperSize,
	}
}

// 从待执行列表弹出一笔订单
// 数据从链表的左边出，右边入
func (k *OrderKeeper) pop() (*model.TransferParams, error) {
	value, err := redis.Client.ListPop(orderKeeperListKey)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("nil")
	}

	transferParams := &model.TransferParams{}
	err = json.Unmarshal(value, transferParams)
	return transferParams, err
}

// 尝试将订单数据推入待执行列表
// 首先判断是否已达到最大处理数量
// 如果订单已存在缓存内，直接返回不处理
// 如果不在，加入到待处理列表
func (k *OrderKeeper) pushIfNotExist(param *model.TransferParams) error {
	log.Infof("尝试将订单放入待执行列表outerOrderNo %s", param.OuterOrderNo)
	listLen, err := redis.Client.ListLen(orderKeeperListKey)
	if err != nil {
		log.Errorf("从缓存获取待执行列表长度出现异常:%v", err)
		return err
	}
	log.Infof("目前待处理列表数量 %d", listLen)
	if listLen >= k.keeperSize {
		log.Infof("当前待处理订单数已达到最大数量(%d)", k.keeperSize)
		return errors.New(fmt.Sprintf("the maximum number(%d) of processes has been reached", k.keeperSize))
	}

	buf, err := json.Marshal(param)
	if err != nil {
		log.Errorf("json Marshal err: %v", err)
		return err
	}
	success, err := redis.Client.SetNX(getKey(param.OuterOrderNo), buf, k.againstReplayCacheExpiration)
	if err != nil {
		log.Errorf("执行redis SetNX命令出现异常:%v", err)
		return err
	}
	if !success {
		log.Infof("订单 %s 在待执行列表已存在", param.OuterOrderNo)
		return errors.New(fmt.Sprintf("outerOrderNo：%s already exist", param.OuterOrderNo))
	}

	// 保存到待执行列表
	err = redis.Client.ListRPush(orderKeeperListKey, buf)
	if err != nil {
		return err
	}
	log.Infof("订单推入完成 %s", param.OuterOrderNo)
	return nil
}

func (k *OrderKeeper) removeFromAgainstReplay(outerOrderNo string) {
	if err := redis.Client.Del(getKey(outerOrderNo)); err != nil {
		log.Error("删除redis缓存失败:%v", err)
	}
}

func (k *OrderKeeper) delProcessedKey(outerOrderNo string) error {
	// 遍历待执行列表，查看订单是否在列表内
	// 如果在列表内，不允许删除key
	lRange, err := redis.Client.ListRange(orderKeeperListKey, 0, 50000)
	if err != nil {
		return err
	}
	for _, o := range lRange {
		transferParams := &model.TransferParams{}
		err := json.Unmarshal([]byte(o), transferParams)
		if err != nil {
			return err
		}
		if transferParams.OuterOrderNo == outerOrderNo {
			return fmt.Errorf("order not completed")
		}
	}

	k.removeFromAgainstReplay(getKey(outerOrderNo))
	return nil
}

func getKey(k string) string {
	return orderKeeperAgainstReplay + k
}
