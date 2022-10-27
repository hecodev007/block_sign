package controller

import (
	conf "custody-merchant-admin/config"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/model/base"
	ModelBase "custody-merchant-admin/model/base"
	"custody-merchant-admin/module/blockChainsApi"
	"custody-merchant-admin/router/web/handler"
	"fmt"
	"strings"
)

// FindAssetsByCoinList
// 币种资产查询接口
func FindAssetsByCoinList(c *handler.Context) error {
	req := new(domain.GetAddrInfo)
	as := new(domain.AssetsSelect)
	req.Chain = c.QueryParam("chain")
	req.Coin = c.QueryParam("coin")
	req.ClientId = c.QueryParam("client_id")
	req.Sign = c.QueryParam("sign")
	as.Offset, as.Limit = c.OffsetPage()
	// 判断用户密钥
	// 判断商户的业务线、币种、用户Id数据
	merchantChains, err := service.GetBindInfoByClientId(req.ClientId)
	if err != nil {
		return handler.OutCodeError(c, 10328, err.Error())
	}
	sinfo, err := service.FirstServiceBySId(int(merchantChains.BusinessId))
	as.MerchantId = sinfo.AccountId
	as.ServiceId = int(merchantChains.BusinessId)
	// 查询币种
	coin, err := base.FindCoinsByChainName(req.Coin, req.Chain)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	if coin.Id == 0 {
		return handler.OutCodeError(c, 30006, "主链币的币种暂无")
	}
	as.CoinState = 0
	as.CoinId = int(coin.Id)
	as.CoinName = req.Coin
	list, total, err := service.FindServiceAssetsList(as, as.MerchantId)
	dataMap := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	return handler.OutResult(c, 10000, "success", dataMap)
}

// GenerateBatchAddress
// 批量创建地址
func GenerateBatchAddress(c *handler.Context) error {
	ba := new(domain.BatchAddrStruct)
	err := c.DataBinder(ba)
	if err != nil {
		return handler.OutCodeError(c, 10305, err.Error())
	}
	req := domain.BatchAddrTo(ba)
	req.Nums = len(req.UserId)
	if req.Nums == 0 {
		return handler.OutCodeError(c, 10305, "user_id为空")
	}
	if req.ClientId == "" {
		return handler.OutCodeError(c, 10305, "client_id为空")
	}
	if req.Sign == "" {
		return handler.OutCodeError(c, 10305, "sign为空")
	}
	if req.Coin == "" {
		return handler.OutCodeError(c, 10305, "coin为空")
	}
	if req.Chain == "" {
		return handler.OutCodeError(c, 10305, "chain为空")
	}
	coin, err := base.FindCoinsByChainName(req.Coin, req.Chain)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	if coin.Id == 0 {
		return handler.OutCodeError(c, 30006, "主链币的币种暂无")
	}

	// 判断用户密钥
	// 判断商户的业务线、币种、用户Id数据
	merchantChains, err := service.GetBindInfoByClientId(req.ClientId)
	if err != nil {
		return handler.OutCodeError(c, 10328, err.Error())
	}
	sinfo, err := service.FirstServiceBySId(int(merchantChains.BusinessId))
	if err != nil {
		return handler.OutCodeError(c, 10325, err.Error())
	}
	if sinfo.Id == 0 {
		return handler.OutCodeError(c, 10325, "没有业务线")
	}

	if !conf.EnvPro {
		verify, err := blockChainsApi.BlockChainVerifyParamFromCustody(ba, conf.Conf.BlockchainCustody.ClientId, conf.Conf.BlockchainCustody.ApiSecret)
		if err != nil {
			return handler.OutCodeError(c, 10305, err.Error())
		}
		if !verify {
			return handler.OutCodeError(c, 10305, "参数校验无法通过")
		}
	}
	err = service.SumAddrNums(int64(req.Nums), merchantChains.BusinessId)
	if err != nil {
		return handler.OutCodeError(c, 10326, err.Error())
	}
	// 调用钱包地址生成接口
	addrList, err := service.CreateBatchChainAddress(merchantChains.BusinessId, req.Coin, req.Nums)
	if err != nil {
		return handler.OutCodeError(c, 10326, err.Error())
	}
	mapList := make([]map[string]interface{}, 0)
	for i, addr := range addrList {
		uid := "0"
		if len(req.UserId) != 0 && len(req.UserId) > i {
			uid = req.UserId[i]
		}
		addrInfo := new(domain.InsertAddrInfo)
		addrInfo.MerchantUser = uid
		addrInfo.ServiceId = merchantChains.BusinessId
		addrInfo.CoinId = int(coin.Id)
		addrInfo.ChainId = coin.ChainId
		addrInfo.MerchantId = sinfo.AccountId
		addrInfo.Address = addr
		// 存储生成的用户地址
		err = service.CreateUserAddress(addrInfo)
		if err != nil {
			return handler.OutCodeError(c, 10325, err.Error())
		}
		mapList = append(mapList, map[string]interface{}{
			"userId":  uid,
			"address": addr,
		})
	}
	// 计入service_combo地址套餐
	service.UpAddrComboUse(int64(len(mapList)), merchantChains.BusinessId)
	return handler.OutResult(c, 10000, "success", mapList)
}

// GenerateUserAddress
// 根据给用户创建单个地址
func GenerateUserAddress(c *handler.Context) error {
	req := new(domain.GetAddrInfo)
	req.UserId = c.QueryParam("user_id")
	req.Coin = c.QueryParam("coin")
	req.Chain = c.QueryParam("chain")
	req.ClientId = c.QueryParam("client_id")
	req.Sign = c.QueryParam("sign")
	req.Nonce = c.QueryParam("nonce")
	req.Ts = c.SwitchType(c.QueryParam("ts"), "int64").(int64)
	//req.SecureKey = c.QueryParam("secureKey")

	if req.UserId == "" {
		return handler.OutCodeError(c, 10305, "user_id为空")
	}
	if req.ClientId == "" {
		return handler.OutCodeError(c, 10305, "client_id为空")
	}
	if req.Sign == "" {
		return handler.OutCodeError(c, 10305, "sign为空")
	}
	if req.Coin == "" {
		return handler.OutCodeError(c, 10305, "coin为空")
	}
	if req.Chain == "" {
		return handler.OutCodeError(c, 10305, "chain为空")
	}
	// 获取币种
	coin, err := base.FindCoinsByName(req.Coin)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	// 获取币种
	chain, err := base.GetChainByCId(int(coin.Id), req.Chain)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	if chain == nil || chain.Id == 0 {
		return handler.OutCodeError(c, 30006, "错误的主链或者币种")
	}
	// 判断用户密钥
	// 判断商户的业务线、币种、用户Id数据
	merchantChains, err := service.GetBindInfoByClientId(req.ClientId)
	if err != nil {
		return handler.OutCodeError(c, 10328, err.Error())
	}
	addr := new(domain.InsertAddrInfo)
	addr.ServiceId = merchantChains.BusinessId
	sinfo, err := service.FirstServiceBySId(int(addr.ServiceId))
	if err != nil {
		return handler.OutCodeError(c, 10329, err.Error())
	}
	if sinfo.Id == 0 {
		return handler.OutCodeError(c, 10329, "没有业务线")
	}
	if !conf.EnvPro {
		verify, err := blockChainsApi.BlockChainVerifyParamFromCustody(req, conf.Conf.BlockchainCustody.ClientId, conf.Conf.BlockchainCustody.ApiSecret)
		if err != nil {
			return handler.OutCodeError(c, 10305, err.Error())
		}
		if !verify {
			return handler.OutCodeError(c, 10305, "参数校验无法通过")
		}
	}
	addr.CoinId = int(coin.Id)
	addr.ChainId = chain.Id
	addr.MerchantId = sinfo.AccountId
	err = service.SumAddrNums(1, addr.ServiceId)
	if err != nil {
		return handler.OutCodeError(c, 10326, err.Error())
	}
	// 调用钱包地址生成接口
	addrList, err := service.CreateBatchChainAddress(merchantChains.BusinessId, req.Coin, 1)
	if err != nil {
		return handler.OutCodeError(c, 10326, err.Error())
	}
	addr.Address = addrList[0]
	// 存储生成的用户地址
	err = service.CreateUserAddress(addr)
	if err != nil {
		return handler.OutCodeError(c, 10325, err.Error())
	}
	service.UpAddrComboUse(1, merchantChains.BusinessId)
	return handler.OutResult(c, 10000, "success", addrList)
}

func CreateWithdraw(c *handler.Context) error {
	req := new(domain.WithdrawStruct)

	// 提现业务线
	err := c.DataBinder(req)
	if err != nil {
		return handler.OutCodeError(c, 10305, err.Error())
	}

	u := domain.StructToInfo(req)
	if u.Amount.IsZero() {
		return handler.OutCodeError(c, 10321, "提现为0")
	}
	if u.FromAddress == "" {
		return handler.OutCodeError(c, 10305, "from_address为空")
	}
	if u.ToAddress == "" {
		return handler.OutCodeError(c, 10305, "to_address为空")
	}
	if u.ClientId == "" {
		return handler.OutCodeError(c, 10305, "client_id为空")
	}
	if u.Sign == "" {
		return handler.OutCodeError(c, 10305, "sign为空")
	}
	if u.Coin == "" {
		return handler.OutCodeError(c, 10305, "coin为空")
	}
	if u.Chain == "" {
		return handler.OutCodeError(c, 10305, "chain为空")
	}
	merchantChains, err := service.GetBindInfoByClientId(u.ClientId)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	u.BusinessId = int(merchantChains.BusinessId)
	//if conf.Conf.Mod == "pro" {
	//
	//}
	fmt.Printf("%v", u)
	params := domain.ParamsToWithdraw(req)
	// 先调用钱包验证签名
	custody, err := blockChainsApi.BlockChainVerifyParamFromCustody(params, conf.Conf.BlockchainCustody.ClientId, conf.Conf.BlockchainCustody.ApiSecret)
	if err != nil {
		return handler.OutCodeError(c, 10305, "签名错误:"+err.Error())
	}
	if !custody {
		return handler.OutCodeError(c, 10305, "参数校验无法通过")
	}
	u.Coin = strings.ToUpper(u.Coin)
	u.Chain = strings.ToUpper(u.Chain)
	// 调用bill记录账单
	cs, err := service.FindMerchantBySidCoin(u.BusinessId, u.Coin)
	if err != nil || cs == nil || cs.CoinId == 0 {
		return handler.OutCodeError(c, 30006, fmt.Sprintf("%s该主链不存在这个代币", u.Chain))
	}
	// 获取币种
	coin, err := ModelBase.FindCoinsByName(u.Coin)
	if err != nil {
		return handler.OutCodeError(c, 30006, err.Error())
	}
	if coin.Id == 0 {
		return handler.OutCodeError(c, 30006, fmt.Sprintf("%s不存在这个代币", u.Coin))
	}
	chains, err := ModelBase.FindChainsById(coin.ChainId)
	if err != nil || chains.Id == 0 {
		return handler.OutCodeError(c, 30006, fmt.Sprintf("%s不存在这个代币", u.Coin))
	}

	w := &domain.BillInfo{}
	w.MerchantId = cs.MerchantId
	w.CreateByUser = cs.MerchantId
	w.TxFromAddr = u.FromAddress
	w.CoinId = cs.CoinId
	w.ChainId = chains.Id
	w.TxToAddr = u.ToAddress
	w.ServiceId = u.BusinessId
	w.Nums = u.Amount
	w.Remark = ""
	w.Memo = u.Memo
	w.FromId = fmt.Sprintf("%d", cs.MerchantId)
	w.State = 0
	// 创建提币订单
	serialNo, err := service.CreateWithdrawBill(w, chains.Name, u.Coin)
	if err != nil {
		return handler.OutCodeError(c, 10327, err.Error())
	}
	return handler.OutResult(c, 10000, "success", map[string]interface{}{"serialNo": serialNo})
}

func FindBillInfos(c *handler.Context) error {
	req := new(domain.GetBillInfo)
	req.SerialNo = c.QueryParam("serial_no")
	req.ClientId = c.QueryParam("client_id")
	req.Sign = c.QueryParam("sign")
	if req.SerialNo == "" {
		return handler.OutCodeError(c, 10305, "serial_no为空")
	}
	if req.ClientId == "" {
		return handler.OutCodeError(c, 10305, "client_id为空")
	}
	if req.Sign == "" {
		return handler.OutCodeError(c, 10305, "sign为空")
	}
	datas, err := service.FindOutBillInfoBySerialNo(req.SerialNo)
	if err != nil {
		return handler.OutCodeError(c, 10327, err.Error())
	}
	return handler.OutResult(c, 10000, "success", datas)
}
