package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/conf"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model/status"
	"github.com/group-coldwallet/blockchains-go/model/transfer"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"xorm.io/builder"
)

const (
	processInterval     = time.Second * 2
	immediatelyInterval = time.Microsecond
	empty               = "nil"
)

type FetchAddressReq struct {
	mchId          int             // 商户id
	outerOrderNo   string          // 外部订单编号
	chain          string          //链
	coin           string          //币种，如果是代币此值为`coinCode`，否则为`chain` ETH/USDT-ERC20
	banFromAddress string          //禁止用来出账的地址
	amount         decimal.Decimal //出账金额
	limitAmount    decimal.Decimal // 限制金额，如果出账的是主链币，需要保留一部分作为手续费
	existCount     int             // 该订单下有效的交易数量
}

type ReplaceRequestData struct {
	OuterOrderNo string                   `json:"outerOrderNo"`
	MulFrom      []TransferRequestMulFrom `json:"mulFrom"`
}

type TransferRequestData struct {
	ApplyId      int64                    `json:"applyId"`
	OuterOrderNo string                   `json:"outerOrderNo"`
	InnerOrderNo string                   `json:"innerOrderNo"`
	Mch          string                   `json:"mch"`
	Chain        string                   `json:"chain"`
	CoinCode     string                   `json:"coinCode"`
	Contract     string                   `json:"contract"`
	Amount       string                   `json:"amount"`
	MulFrom      []TransferRequestMulFrom `json:"mulFrom"`
	ToAddress    string                   `json:"toAddress"`
}

type TransferRequestMulFrom struct {
	FromAddress string `json:"fromAddress"`
	Amount      string `json:"amount"`
}

type TxMultiFromAddrAmount struct {
	Address string
	Amount  decimal.Decimal
	// 该地址原冻结金额，不包含Amount
	FreezeAmount decimal.Decimal
	//ReserveFeeAmount decimal.Decimal
}

type TxManager struct {
	mtx sync.RWMutex

	// 按链进行加锁
	// map[chainName] = sync.RWMutex
	freezeCoinMtx sync.Map

	// 按订单号进行加锁
	// map[outerOrderNo] = sync.RWMutex
	replaceMtx sync.Map
}

func NewTxManager() *TxManager {
	return &TxManager{}
}

func (tm *TxManager) StartAsync(ctx context.Context) {
	go tm.loopProcess(ctx)
}

func (tm *TxManager) loopProcess(ctx context.Context) {
	timer := time.NewTimer(processInterval)
loop:
	for {
		select {
		case <-timer.C:
			if applyId, err := tm.processPopOrder(); err != nil {
				if err.Error() != empty {
					dingding.ErrTransferDingBot.NotifyStr(err.Error())
					dao.FcTransfersApplyUpdateRemark(int64(applyId), err.Error())
					log.Errorf("循环执行待处理订单交易出错 %v", err.Error())
				}
				timer.Reset(processInterval)
			} else {
				timer.Reset(immediatelyInterval)
			}
		case <-ctx.Done():
			timer.Stop()
			break loop
		}
	}
}

func (tm *TxManager) processPopOrder() (int, error) {
	val, err := redis.Client.ListPop(redis.CacheKeyWaitingOrderList)
	if err != nil {
		return 0, err
	}
	if val == nil {
		return 0, errors.New("nil")
	}

	outOrderNo := string(val)
	log.Infof("processPopOrder 准备处理订单 %s", outOrderNo)
	apply, err := dao.FcTransfersApplyByOutOrderNo(outOrderNo)
	if err != nil {
		return apply.Id, err
	}

	priorityApplyId, exist := tm.getPriorityOrderThisChain(apply.CoinName, apply.Eoskey, apply.AppId)
	if exist {
		if priorityApplyId != apply.Id {
			log.Infof("%s 存在需要优先处理的订单applyId=%d,当前订单%s不可执行", apply.CoinName, priorityApplyId, apply.OutOrderid)
			// 重新放入到缓存，等待下一次执行
			order.PushToWaitingList(apply.OutOrderid)
			return 0, errors.New("nil")
		}
	}

	lockCacheKey := redis.GetTxProcessLockCacheKey(outOrderNo, apply.Applicant)
	existLock, err := redis.Client.Get(lockCacheKey)
	if err != nil {
		return apply.Id, err
	}
	if existLock != "" {
		return apply.Id, fmt.Errorf("订单%s已被执行或正在执行，已被锁定", outOrderNo)
	}

	if err = checkIsRunOnOldVersion(apply); err != nil {
		return apply.Id, err
	}
	log.Infof("processPopOrder 检查订单是否在旧版本执行通过 %s", outOrderNo)
	if err = checkIsApplyStatus(apply.OutOrderid, apply.Status); err != nil {
		return apply.Id, err
	}
	log.Infof("processPopOrder 检查订单transfer_apply当前状态通过 %s", outOrderNo)
	if err = redis.Client.Set(lockCacheKey, outOrderNo, time.Hour*24*14); err != nil {
		return apply.Id, err
	}
	go tm.process(apply)
	return 0, nil
}

func (tm *TxManager) process(apply *entity.FcTransfersApply) {
	if err := tm.doTx(apply); err != nil {
		log.Infof("处理多地址出账交易失败 %v", err)
		dingding.ErrTransferDingBot.NotifyStr(fmt.Sprintf("订单: %s\n%s", apply.OutOrderid, err.Error()))
		dao.FcTransfersApplyUpdateRemark(int64(apply.Id), err.Error())
		redis.Client.Del(redis.GetTxProcessLockCacheKey(apply.OutOrderid, apply.Applicant))
	}
}

func (tm *TxManager) doTx(apply *entity.FcTransfersApply) error {
	if err := checkIfExistOrders(apply.OutOrderid, apply.CoinName, apply.AppId); err != nil {
		return fmt.Errorf("[检查订单是否重复失败] %v", err)
	}
	log.Infof("processPopOrder 检查订单是否存在通过 %s", apply.OutOrderid)
	if err := checkIfInvalidChain(apply.CoinName); err != nil {
		return fmt.Errorf("[检查链是否有效失败] %v", err)
	}
	if err := checkOrderSecure(apply.Id, apply.OutOrderid); err != nil {
		return fmt.Errorf("[订单安全检查失败] %v", err)
	}
	log.Infof("processPopOrder 检查订单安全信息通过 %s", apply.OutOrderid)

	pack, err := tm.buildRequestOrderData(apply)
	if err != nil {
		return fmt.Errorf("[多地址出账构建数据失败] %v", err)
	}
	callUrl := fmt.Sprintf("%s/v2/transfer", conf.Cfg.Walletserver.Url)
	err = tm.callWalletServer(callUrl, pack)

	newStatus := entity.ApplyStatus_CreateOk
	reason := ""
	if err != nil {
		tm.unlockFreeze(pack.Chain, pack.CoinCode, pack.OuterOrderNo, pack.MulFrom)
		log.Errorf("多地址出账调用walletServer失败 %v", err)
		reason = fmt.Sprintf("多地址出账调用walletServer失败:%v", err)
		newStatus = entity.ApplyStatus_Creating // 多地址出账，43才能重推
	}
	if errUpdateSta := dao.FcTransfersApplyUpdateStatusAndRemarkById(apply.Id, int(newStatus), reason); errUpdateSta != nil {
		return fmt.Errorf("多地址出账更新transfers_apply状态失败(提示：订单金额已冻结) %v", errUpdateSta)
	}
	return err
}

func (tm *TxManager) unlockFreeze(chain, coinCode, outerOrderNo string, mulFrom []TransferRequestMulFrom) {
	store, _ := tm.freezeCoinMtx.LoadOrStore(chain, &sync.RWMutex{})
	mtx := store.(*sync.RWMutex)
	defer mtx.Unlock()
	mtx.Lock()

	log.Infof("解冻订单%s 数据 %+v", outerOrderNo, mulFrom)
	coinType := chain
	if coinCode != "" {
		coinType = coinCode
	}
	models := make([]dao.FcUpdateFreeze, 0)
	for _, f := range mulFrom {
		models = append(models, dao.FcUpdateFreeze{Address: f.FromAddress, FreezeAmount: f.Amount})
	}
	if err := dao.FcAddressAmountUpdateUnFreeze(coinType, models); err != nil {
		msg := fmt.Sprintf("执行构建订单交易(%s)失败时尝试解冻金额出错:%v", outerOrderNo, err)
		log.Infof(msg)
		dingding.WarnDingBot.NotifyStr(msg)
	}
	log.Infof("解冻订单%s已完成", outerOrderNo)
}

func (tm *TxManager) callWalletServer(url string, data interface{}) error {
	log.Infof("准备调用walletServer %s，请求参数=%v", url, data)
	resp, err := util.PostJsonByAuth(url, conf.Cfg.Walletserver.User, conf.Cfg.Walletserver.Password, data)
	if err != nil {
		return fmt.Errorf("调用walletServer出错: %v", err)
	}

	log.Infof("%s 调用walletServer返回数据 %s", url, string(resp))
	result := transfer.DecodeWalletServerRespOrder(resp)
	if !result.Success() {
		return fmt.Errorf("call walletServer failure:%v", result.Message)
	}
	return nil
}

func (tm *TxManager) buildRequestOrderData(apply *entity.FcTransfersApply) (*TransferRequestData, error) {
	log.Infof("processPopOrder 准备构建订单 %+v", apply)
	mch, err := dao.FcMchFindByPlatform(apply.Applicant)
	if err != nil {
		return nil, fmt.Errorf("根据商户编号(%s)查询商户数据失败 %v", apply.Applicant, err)
	}

	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id, "address_flag": "to"})
	if err != nil {
		return nil, err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return nil, fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", apply.Id, apply.OutOrderid)
	}
	toEntity := toAddrs[0]

	reqData := &TransferRequestData{
		ApplyId:      int64(apply.Id),
		OuterOrderNo: apply.OutOrderid,
		InnerOrderNo: apply.OrderId,
		Mch:          apply.Applicant,
		Chain:        apply.CoinName,
		Amount:       toEntity.ToAmount,
		ToAddress:    toAddrs[0].Address,
	}

	limitAmount := decimal.Zero
	coinSet := &entity.FcCoinSet{}

	isTokenTransfer := apply.Eostoken != ""
	coinType := ""
	if isTokenTransfer {
		coinSet = global.CoinDecimal[apply.Eoskey]
		if coinSet == nil {
			return nil, fmt.Errorf("读取 %s coinSet 设置异常", apply.Eoskey)
		}
		if strings.ToLower(coinSet.Token) != strings.ToLower(apply.Eostoken) {
			return nil, fmt.Errorf("合约地址不匹配 %s <> %s", coinSet.Token, apply.Eostoken)
		}
		reqData.CoinCode = apply.Eoskey
		reqData.Contract = apply.Eostoken
		coinType = reqData.CoinCode
	} else {
		coinSet = global.CoinDecimal[reqData.Chain]
		limitAmount, _ = decimal.NewFromString(coinSet.StaThreshold)
		coinType = reqData.Chain
	}
	//precision := coinSet.Decimal
	toAmtDecimal, _ := decimal.NewFromString(reqData.Amount)
	log.Infof("processPopOrder 准备挑选合适的出账地址 %s", apply.OutOrderid)

	far := FetchAddressReq{
		mchId:          mch.Id,
		outerOrderNo:   apply.OutOrderid,
		banFromAddress: toEntity.BanFromAddress,
		chain:          reqData.Chain,
		coin:           coinType,
		amount:         toAmtDecimal,
		limitAmount:    limitAmount,
		existCount:     0,
	}

	txMultiList, err := tm.fetchFromAddressAndFreezeWithLock(far)
	if err != nil {
		return nil, fmt.Errorf("挑选合适的出账地址和金额失败:%v", err)
	}

	mulFrom := make([]TransferRequestMulFrom, 0)
	for _, t := range txMultiList {
		mulFrom = append(mulFrom, TransferRequestMulFrom{
			FromAddress: t.Address,
			Amount:      t.Amount.String(),
		})
	}

	reqData.MulFrom = mulFrom
	return reqData, nil
}

// fetchFromAddressAndFreezeWithLock 按链加锁
// 挑选合适的出账地址和金额，并冻结金额
func (tm *TxManager) fetchFromAddressAndFreezeWithLock(req FetchAddressReq) ([]TxMultiFromAddrAmount, error) {
	mchId := req.mchId
	chain := req.chain
	coin := req.coin
	amount := req.amount
	limitAmount := req.limitAmount
	banFromAddress := req.banFromAddress

	store, _ := tm.freezeCoinMtx.LoadOrStore(chain, &sync.RWMutex{})
	mtx := store.(*sync.RWMutex)
	defer mtx.Unlock()
	mtx.Lock()

	log.Infof("fetchFromAddressAndFreezeWithLock coin=%s，mchId=%d 需要出账金额%s，限制最小金额%s", coin, mchId, amount.String(), limitAmount.String())
	limit := conf.Cfg.MultiLimit
	if limit == 0 {
		limit = 5
	}
	list, err := dao.FcAddressAmountExcludeFreeze(mchId, limitAmount.String(), coin, limit+2)
	if err != nil {
		return nil, err
	}
	log.Infof("fetchFromAddressAndFreezeWithLock 从fc_address_amount获取到数据条数为 %+v", list)

	if len(list) == 0 {
		return nil, fmt.Errorf("币种:%s 出账金额不足，需要出账:%s，合适的出账的地址为0", coin, amount.String())
	}

	models := make([]TxMultiFromAddrAmount, 0)

	// 用户地址 + 出账地址
	availableList := make([]dao.FcExcludeFreeze, 0)
	// 用户地址
	userAddrs := make([]dao.FcExcludeFreeze, 0)
	for _, l := range list {
		if strings.ToLower(l.Address) == strings.ToLower(banFromAddress) {
			log.Infof("fetchFromAddressAndFreezeWithLock 当前挑选的地址 %s 是 指定不可用来出账的地址，忽略", l.Address)
			continue
		}
		availableList = append(availableList, l)

		if l.Type == 2 {
			userAddrs = append(userAddrs, l)
		}
	}
	if len(availableList) == 0 {
		//cltMsg := "success"
		//if err = tm.doCollect([]string{banFromAddress}, req.outerOrderNo, req.contract, coin, amount); err != nil {
		//	cltMsg = err.Error()
		//}
		return nil, fmt.Errorf("\n币种:%s\n排除不可用地址(%s)后出账金额不足\n需要出账:%s\n无其他可用的出账地址\n请使用命令：订单归集%s，进行归集后出账", coin, banFromAddress, amount.String(), req.outerOrderNo)
	}

	// 优先使用用户地址来出账
	// 如果用户地址凑钱出账失败，再尝试把
	models, err = tm.fetchCore(userAddrs, req, limit)
	if err != nil {
		log.Infof("fetchFromAddressAndFreezeWithLock 尝试只挑选用户地址出账失败 %v", err)
		log.Info("fetchFromAddressAndFreezeWithLock 挑选用户地址 + 出账地址出账")
		models, err = tm.fetchCore(availableList, req, limit)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("fetchFromAddressAndFreezeWithLock 挑选用户地址出账成功")
	}

	if len(models)+req.existCount > limit {
		return nil, fmt.Errorf("每笔订单限制交易数量:%d\n该订单已存在有效的交易数:%d\n本次出账需要使用%d笔，超过限制数量，建议归集后再出账", limit, req.existCount, len(models))
	}

	log.Info("fetchFromAddressAndFreezeWithLock 准备冻结金额")
	if err = tm.freezeAddressAndAmount(coin, models); err != nil {
		return nil, err
	}
	log.Info("fetchFromAddressAndFreezeWithLock 冻结金额完成")
	return models, nil
}

func (tm *TxManager) fetchCore(list []dao.FcExcludeFreeze, req FetchAddressReq, limit int) ([]TxMultiFromAddrAmount, error) {
	limitAmount := req.limitAmount
	amount := req.amount
	coin := req.coin

	if len(list) == 0 {
		return nil, errors.New("没有可出账的地址信息")
	}
	cmpTotal := decimal.Zero
	models := make([]TxMultiFromAddrAmount, 0)
	for _, item := range list {
		actualTakeAmt := decimal.Zero // 地址实际出账的金额
		estimateFee := limitAmount    // 预估的手续费，仅当是主链币转账才会有值
		thisAmt, _ := decimal.NewFromString(item.Amount)
		thisAmt = tm.scaled(thisAmt).Sub(estimateFee)
		// 差额 = 需要出账的金额 - 已统计的总金额
		diff := amount.Sub(cmpTotal)
		if thisAmt.LessThan(diff) {
			// 当前地址金额 < 差额
			actualTakeAmt = thisAmt
		} else {
			// 当前地址金额 >= 差额
			actualTakeAmt = diff
		}

		cmpTotal = cmpTotal.Add(actualTakeAmt)
		freezeDecimal, _ := decimal.NewFromString(item.FreezeAmount)
		models = append(models, TxMultiFromAddrAmount{Address: item.Address, Amount: actualTakeAmt, FreezeAmount: freezeDecimal})
		if cmpTotal.GreaterThanOrEqual(amount) {
			break
		}
	}
	if len(models) > limit {
		return nil, fmt.Errorf("%s获取前%d个地址总额为%s，不足以出账", coin, limit, cmpTotal.String())
	}
	if cmpTotal.GreaterThan(amount) {
		return nil, fmt.Errorf("%s获取前%d个地址总额为%s，超出出账（需要金额%s）", coin, limit, cmpTotal.String(), amount.String())
	}
	if cmpTotal.LessThan(amount) {
		//ifNeedCollect, err := order.CheckIfNeedCollect(req.mchId, req.outerOrderNo, req.chain, req.CoinCode, req.contract, req.collectThreshold, req.amount)
		banMsg := ""
		if req.banFromAddress != "" {
			banMsg = fmt.Sprintf("\n本次要求排除的出账地址：%s", req.banFromAddress)
		}
		return nil, fmt.Errorf("\n%s获取前%d个地址总额为%s\n不足以出账（需要金额%s）%s\n请使用命令：订单归集%s，进行归集后出账", coin, limit, cmpTotal.String(), amount.String(), banMsg, req.outerOrderNo)
	}
	if err := checkIfAmountEnough(amount, models); err != nil {
		return nil, err
	}
	return models, nil
}

// scaled 计算金额的时候取7位小数
func (tm *TxManager) scaled(val decimal.Decimal) decimal.Decimal {
	precision := int32(7)
	str := val.Shift(precision).StringScaled(0)
	fs, _ := decimal.NewFromString(str)
	return fs.Shift(-precision)
}

func (tm *TxManager) freezeAddressAndAmount(coinType string, models []TxMultiFromAddrAmount) error {
	entities := make([]dao.FcUpdateFreeze, 0)
	for _, m := range models {
		entities = append(entities, dao.FcUpdateFreeze{Address: m.Address, FreezeAmount: m.Amount.String()})
	}
	return dao.FcAddressAmountUpdateFreeze(coinType, entities)
}

func (tm *TxManager) doCollect(address []string, outerOrderNo, contract, coin string, amount decimal.Decimal) error {
	return order.CallCollectCenter(outerOrderNo, address, contract, coin, amount)
}

func (tm *TxManager) getOrderTotalAmount(outerOrderNo string) (decimal.Decimal, error) {
	amount := decimal.Zero
	orders, err := dao.FcOrderFindListByOutNo(outerOrderNo)
	if err != nil {
		return amount, err
	}
	if len(orders) == 1 {
		amount, _ = decimal.NewFromString(orders[0].TotalAmount)
		return amount, nil
	}

	orderHots, err := dao.FcOrderHotFindByOutNo(outerOrderNo)
	if err != nil {
		return amount, err
	}
	if len(orderHots) == 1 {
		amount, _ = decimal.NewFromString(orderHots[0].TotalAmount)
		return amount, nil
	}
	return amount, errors.New("order not found")
}

func (tm *TxManager) ReplaceTxs(outerOrderNo string) error {
	store, _ := tm.freezeCoinMtx.LoadOrStore(outerOrderNo, &sync.RWMutex{})
	mtx := store.(*sync.RWMutex)
	defer mtx.Unlock()
	mtx.Lock()

	validList, err := dao.FindOrderTxsByOuterOrderNo(outerOrderNo)
	log.Infof("替换失败交易 查询到交易记录 %+v", validList)
	if err != nil {
		return err
	}
	if len(validList) == 0 {
		return fmt.Errorf("没有找到可替换的交易，请将链上失的交易设置为[12 链上失败]")
	}

	orderTotalAmount, err := tm.getOrderTotalAmount(outerOrderNo)
	if err != nil {
		return fmt.Errorf("查询订单总金额出错 %v", err)
	}

	mch := validList[0].Mch
	chain := validList[0].Chain
	coinCode := validList[0].CoinCode

	// 有效的金额（排除链上失败、已取消）
	totalValidAmt := decimal.Zero
	totalValidCount := 0
	for _, tx := range validList {
		if tx.IsChainFailure() {
			continue
		}
		if tx.IsCanceled() {
			continue
		}
		fs, _ := decimal.NewFromString(tx.Amount)
		totalValidAmt = totalValidAmt.Add(fs)
		totalValidCount += 1
	}

	// 本次需要替换的金额为，订单总金额 - 有效交易金额
	replaceAmount := orderTotalAmount.Sub(totalValidAmt)

	mchModel, err := dao.FcMchFindByPlatform(mch)
	if err != nil {
		return fmt.Errorf("根据商户编号(%s)查询商户数据失败 %v", mch, err)
	}

	limitAmount := decimal.Zero
	coinType := coinCode
	coinSet := &entity.FcCoinSet{}
	if coinCode == "" {
		coinSet = global.CoinDecimal[chain]
		limitAmount, _ = decimal.NewFromString(coinSet.StaThreshold)
		coinType = chain
	} else {
		coinSet = global.CoinDecimal[coinCode]
	}
	//precision := coinSet.Decimal
	log.Infof("替换失败交易 准备挑选合适的地址出账 %s %s %s %s", chain, coinType, replaceAmount.String(), limitAmount)

	banFromAddress, err := getBanFromAddress(outerOrderNo)
	if err != nil {
		return fmt.Errorf("根据订单号%s获取禁止使用的出账地址失败 %v", outerOrderNo, err)
	}

	far := FetchAddressReq{
		mchId:          mchModel.Id,
		outerOrderNo:   outerOrderNo,
		banFromAddress: banFromAddress,
		chain:          chain,
		coin:           coinType,
		amount:         replaceAmount,
		limitAmount:    limitAmount,
		existCount:     totalValidCount,
	}
	txMultiList, err := tm.fetchFromAddressAndFreezeWithLock(far)
	if err != nil {
		return fmt.Errorf("订单：%s\n替换失败交易-挑选合适的出账地址和金额失败:\n%v", outerOrderNo, err)
	}
	log.Infof("替换失败交易 已挑选地址 %+v", txMultiList)

	mulFrom := make([]TransferRequestMulFrom, 0)
	for _, t := range txMultiList {
		mulFrom = append(mulFrom, TransferRequestMulFrom{
			FromAddress: t.Address,
			Amount:      t.Amount.String(),
		})
	}

	reqData := &ReplaceRequestData{
		OuterOrderNo: outerOrderNo,
		MulFrom:      mulFrom,
	}
	callUrl := fmt.Sprintf("%s/v2/replacetx", conf.Cfg.Walletserver.Url)
	err = tm.callWalletServer(callUrl, reqData)
	if err != nil {
		tm.unlockFreeze(chain, coinCode, outerOrderNo, mulFrom)
		errMsg := fmt.Sprintf("替换失败交易调用walletServer失败 %v", err)
		log.Errorf(errMsg)
		dingding.WarnDingBot.NotifyStr(errMsg)
		return errors.New(errMsg)
	}
	log.Infof("替换失败交易 完成")
	return nil
}

// getPriorityOrderThisChain 获取这个链尚未处理的优先订单
// chain 链名
// return applyId,是否存在
func (tm *TxManager) getPriorityOrderThisChain(chain, coinCode string, mchId int) (int, bool) {
	priorities, err := dao.FcOrderPriorityByChain(chain, coinCode, mchId)
	if err != nil {
		log.Infof("执行getPriorityOrderThisChain 出错：%v", err)
		return 0, false
	}
	if len(priorities) == 0 {
		return 0, false
	}
	log.Infof("链=%s 存在需要优先处理的订单=%s", chain, priorities[0].OuterOrderNo)
	return priorities[0].ApplyId, true
}

func getBanFromAddress(outerOrderNo string) (string, error) {
	apply, err := dao.FcTransfersApplyByOutOrderNo(outerOrderNo)
	if err != nil {
		return "", err
	}

	//查询出账地址和金额
	toAddrs, err := entity.FcTransfersApplyCoinAddress{}.Find(builder.Eq{"apply_id": apply.Id, "address_flag": "to"})
	if err != nil {
		return "", err
	}
	//一般出账地址只有一个
	if len(toAddrs) != 1 {
		return "", fmt.Errorf("内部订单ID：%d，外部订单号：%s,接受地址只允许一个", apply.Id, apply.OutOrderid)
	}
	return toAddrs[0].BanFromAddress, nil
}

func checkIfAmountEnough(amount decimal.Decimal, models []TxMultiFromAddrAmount) error {
	total := decimal.Zero
	for _, m := range models {
		total = total.Add(m.Amount)
	}
	if !total.Equal(amount) {
		return fmt.Errorf("计算得到的金额不相等：%s != %s", amount.String(), total.String())
	}
	return nil
}

func checkOrderSecure(applyId int, outOrderNo string) error {
	err := validTransferApplyBase(applyId)
	if err != nil {
		msg := fmt.Sprintf("入侵订单，outOrderId:%s,applyId:%d, error:%s", outOrderNo, applyId, err.Error())
		dingding.WarnDingBot.NotifyStr(msg)
		return errors.New(msg)
	}
	return nil
}

func checkIfInvalidChain(chain string) error {
	_, has := transferService[strings.ToLower(chain)]
	if !has {
		return fmt.Errorf("缺少相关币种服务初始化 ==> %s", coinName)
	}
	return nil
}

func checkIfExistOrders(outOrderNo, chain string, mchId int) error {
	if global.WalletType(chain, mchId) == status.WalletType_Cold {
		list, err := dao.FcOrderFindListByOutNo(outOrderNo)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			return fmt.Errorf("%s already exist in fc_order", outOrderNo)
		}
	} else {
		list, err := dao.FcOrderHotFindByOutNo(outOrderNo)
		if err != nil {
			return err
		}
		if len(list) > 0 {
			return fmt.Errorf("%s already exist in fc_order_hot", outOrderNo)
		}
	}
	return nil
}

func checkIsRunOnOldVersion(apply *entity.FcTransfersApply) error {
	redisHelper, _ := util.AllocRedisClient()
	defer redisHelper.Close()

	if entity.MultiAddrTx != apply.TxType {
		return fmt.Errorf("checkIsRunOnOldVersion 订单%s出账类型不正确", apply.OutOrderid)
	}
	interceptKey := fmt.Sprintf("%d_%s", apply.AppId, apply.OutOrderid)
	existRunning, _ := redisHelper.Get(interceptKey)
	if existRunning != "" {
		return fmt.Errorf("checkIsRunOnOldVersion 订单%s正在旧版本执行", apply.OutOrderid)
	}
	return nil
}

func checkIsApplyStatus(outOrderNo string, status int) error {
	if int(entity.ApplyStatus_Creating) != status {
		return fmt.Errorf("checkIsApplyStatus 订单%s状态不正确", outOrderNo)
	}
	return nil
}
