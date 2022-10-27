package job

//tip:php原有冷钱包业务需要 wallet去回调，，现在更改流程，自己去找（后续可以通过队列消费）
//预留,暂时不使用

// Job Specific Functions
type TransferOrderCheckJob struct {
}

func (e TransferOrderCheckJob) Run() {
	//log.Info("=======查询需要检查的交易，尚未超过重试次数=======")
	////查询数据,每次查询10条
	//results, err := dao.FcTransfersApplyFindCreateOrder(10, global.RetryNum)
	//if err != nil {
	//	log.Errorf("查询订单数据异常:%s", err.Error())
	//	return
	//}
	//if len(results) == 0 {
	//	log.Info("暂无交易")
	//	return
	//}
	////数据交易
	//for _, v := range results {
	//	if v.ErrorNum >= global.RetryNum {
	//		log.Errorf("订单：%s,重试次数上限", v.OutOrderid)
	//		continue
	//	}
	//	//查询是否存在相关交易的币种,避免与php冲突
	//	coinName := strings.ToLower(v.CoinName)
	//	_, ok := transferService[coinName]
	//	if !ok {
	//		log.Errorf("缺少相关币种服务初始化 ==> %s", coinName)
	//		continue
	//	}
	//
	//	//查询order表是否成功
	//	records, err := dao.FcOrderFindByOutNo(v.OutOrderid)
	//	if err != nil {
	//		//理论上不存在空数据
	//		log.Errorf("订单：%s,查询异常,err :%s", v.OutOrderid, err.Error())
	//	}
	//	if len(records) == 0 {
	//		//理论上不存在空数据
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//
	//	//订单排序之后只获取最后一笔校验
	//	record := records[0]
	//	//解密
	//	address, amount, err := util.DecodeOrderId(record.OuterOrderNo, record.OrderNo)
	//	if err != nil {
	//		//无法解析订单ID
	//		log.Errorf("订单：%s 校验异常:%s", v.OutOrderid, err.Error())
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//	amountD, _ := decimal.NewFromString(amount)
	//	if amountD.LessThanOrEqual(decimal.Zero) {
	//		log.Errorf("订单：%s 金额校验异常:%s", v.OutOrderid, amountD.String())
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//
	//	dbAmountD, _ := decimal.NewFromString(record.Amount)
	//	bit := global.CoinDecimal[record.CoinName]
	//	if bit == nil {
	//		log.Errorf("订单检查，缺少精度读取币种：%s", record.CoinName)
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//	dbAmountD = dbAmountD.Shift(int32(bit.Decimal))
	//
	//	if address != record.ToAddress || !amountD.Equals(dbAmountD) {
	//		log.Errorf("订单检查，db数据对比异常，"+
	//			"db-addr:%s,"+
	//			"db-amount:%,"+
	//			"check-addr:%s,"+
	//			"check-amount:%s",
	//			record.ToAddress,
	//			dbAmountD.String(),
	//			address,
	//			amountD.String(),
	//		)
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//
	//	//查询订单出账地址对比
	//	addressInfos, err := dao.FcTransfersApplyCoinAddressFindAddrInfo(record.Id, "to")
	//	if err != nil {
	//		//理论上不存在空数据
	//		log.Errorf("订单：%s,查询地址异常,err :%s", v.OutOrderid, err.Error())
	//	}
	//	if len(addressInfos) == 0 {
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		continue
	//	}
	//	//转换金额对比
	//	toAmountD, _ := decimal.NewFromString(addressInfos[0].ToAmount)
	//	if address != addressInfos[0].Address || !amountD.Equals(toAmountD) {
	//		//订单校验无法通过
	//		//钉钉通知人工处理，不然会一直卡住这笔
	//		log.Errorf("订单检查，原始数据对比异常，"+
	//			"dbay-addr:%s,"+
	//			"dbay-amount:%,"+
	//			"check-addr:%s,"+
	//			"check-amount:%s",
	//			addressInfos[0].Address,
	//			toAmountD.String(),
	//			address,
	//			amountD.String(),
	//		)
	//		continue
	//	}
	//
	//	if record.Status < int(status.BroadcastStatus) {
	//		//正在执行，无需处理
	//		continue
	//	} else if record.Status == int(status.BroadcastStatus) {
	//		//修改apply状态，通知已经完成,同时通知回调
	//		err = dao.FcTransfersApplyUpdateStatusById(v.Id, int(entity.ApplyStatus_TransferOk))
	//		if err != nil {
	//			log.Errorf("订单检查错误，修改状态异常，outorderid::%s", v.OutOrderid)
	//			//钉钉通知人工处理，不然会一直卡住这笔
	//			continue
	//		}
	//		//进入通知
	//		orderService.NotifyToMch(v)
	//	} else if record.Status == int(status.BroadcastErrorStatus) || record.Status == int(entity.UnknowErrorStatus) {
	//
	//	}
	//
	//}
}
