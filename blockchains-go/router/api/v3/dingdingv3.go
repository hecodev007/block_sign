package v3

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/model"
	dingModel "github.com/group-coldwallet/blockchains-go/model/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"strings"
	"time"
)

func SetPriorityOrder(content string) error {
	req := &model.DingPriorityRequest{}
	jsonDataStr := strings.Replace(content, dingModel.DING_SET_PRIORITY_ORDER.ToString(), "", -1)
	if err := json.Unmarshal([]byte(jsonDataStr), req); err != nil {
		return err
	}
	if req.OuterOrderNo == "" {
		return errors.New("订单号不能为空")
	}

	apply, err := dao.FcTransfersApplyByOutOrderNo(req.OuterOrderNo)
	if err != nil {
		return fmt.Errorf("transferApply出错:%v", err)
	}
	if apply.Status == int(entity.ApplyStatus_TransferOk) {
		return fmt.Errorf("订单%s 已完成，不可进行此操作", req.OuterOrderNo)
	}

	existPriority, err := dao.FcOrderPriorityByOuterOrderNo(req.OuterOrderNo)
	if err != nil {
		return err
	}
	if existPriority != nil {
		return fmt.Errorf("订单%s 已被设置为优先，设置时间=%v，状态=%d (1：正在出账，2：出账完成)", existPriority.OuterOrderNo, existPriority.CreateTime, existPriority.Status)
	}

	order := &entity.FcOrderPriority{
		OuterOrderNo: req.OuterOrderNo,
		ApplyId:      apply.Id,
		MchId:        apply.AppId,
		ChainName:    apply.CoinName,
		CoinCode:     apply.Eoskey,
		CreateTime:   time.Now(),
		Status:       entity.OrderPriorityStatusProcessing,
	}
	_, err = order.Add()
	if err != nil {
		return fmt.Errorf("订单%s 设置到优先处理失败 %v", req.OuterOrderNo, err)
	}
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单%s 已设置为优先处理订单", req.OuterOrderNo))
	return nil
}

func CancelPriorityOrder(content string) error {
	req := &model.DingPriorityRequest{}
	jsonDataStr := strings.Replace(content, dingModel.DING_CANCEL_PRIORITY_ORDER.ToString(), "", -1)
	if err := json.Unmarshal([]byte(jsonDataStr), req); err != nil {
		return err
	}
	if req.OuterOrderNo == "" {
		return errors.New("订单号不能为空")
	}
	existPriority, err := dao.FcOrderPriorityByOuterOrderNo(req.OuterOrderNo)
	if err != nil {
		return err
	}
	if existPriority == nil {
		return fmt.Errorf("订单%s 还没有设置优先处理", existPriority.OuterOrderNo)
	}

	if existPriority.Status != entity.OrderPriorityStatusProcessing {
		return fmt.Errorf("订单%s 当前状态（%d）不允许取消设置 (1：正在出账，2：出账完成)", existPriority.OuterOrderNo, existPriority.Status)
	}
	dao.DeletePriorityOrder(existPriority.Id)
	dingding.ReviewDingBot.NotifyStr(fmt.Sprintf("订单%s 已取消优先处理", req.OuterOrderNo))
	return nil
}
