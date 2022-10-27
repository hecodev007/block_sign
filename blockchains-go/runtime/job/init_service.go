package job

import (
	"github.com/group-coldwallet/blockchains-go/runtime/dingding"
	"github.com/group-coldwallet/blockchains-go/service"
	"github.com/group-coldwallet/blockchains-go/service/order"
	"github.com/group-coldwallet/blockchains-go/service/security"
	"github.com/group-coldwallet/blockchains-go/service/transfer"
)

var (
	//币种交易服务
	transferService map[string]service.TransferService

	//相关验证服务
	transferSecurityService service.TransferSecurityService

	orderService service.OrderService

	ErrDingBot *dingding.DingBot
)

func init() {
	//币种交易服务
	coinSrv := make(map[string]service.TransferService, 0)
	coinSrv["btc"] = transfer.NewBtcTransferService()
	coinSrv["cocos"] = transfer.NewCocosTransferService()
	coinSrv["mdu"] = transfer.NewMduTransferService()
	coinSrv["zvc"] = transfer.NewZvcTransferService()
	coinSrv["eos"] = transfer.NewEosTransferService()
	coinSrv["eth"] = transfer.NewEthTransferService()
	coinSrv["fo"] = transfer.NewFoTransferService()
	coinSrv["waxp"] = transfer.NewWaxpTransferService()
	coinSrv["klay"] = transfer.NewKlayTransferService()
	coinSrv["gxc"] = transfer.NewGxcTransferService()
	coinSrv["etc"] = transfer.NewEtcTransferService()
	coinSrv["seek"] = transfer.NewSeekTransferService()
	coinSrv["usdt"] = transfer.NewUsdtTransferService()
	coinSrv["kava"] = transfer.NewKavaTransferService()
	coinSrv["luna"] = transfer.NewLunaTransferService()
	coinSrv["lunc"] = transfer.NewLuncTransferService()
	//write by jun 2020/4/29
	coinSrv["bnb"] = transfer.NewBnbTransferService()
	coinSrv["xlm"] = transfer.NewXlmTransferService()
	coinSrv["rub"] = transfer.NewRubTransferService()
	coinSrv["hx"] = transfer.NewHxTransferService()
	coinSrv["cds"] = transfer.NewCdsTransferService()
	coinSrv["ont"] = transfer.NewOntTransferService()
	coinSrv["ar"] = transfer.NewARTransferService()
	coinSrv["ksm"] = transfer.NewKsmTransferService()
	coinSrv["bnc"] = transfer.NewBncTransferService()
	coinSrv["crab"] = transfer.NewCringTransferService()
	coinSrv["hnt"] = transfer.NewHntTransferService()
	coinSrv["vet"] = transfer.NewVetTransferService()
	coinSrv["bsv"] = transfer.NewBsvTransferService()
	coinSrv["ltc"] = transfer.NewLtcTransferService()
	coinSrv["uca"] = transfer.NewUcaTransferService()
	coinSrv["celo"] = transfer.NewCeloTransferService()
	coinSrv["mtr"] = transfer.NewMtrTransferService()
	coinSrv["fio"] = transfer.NewFioTransferService()
	coinSrv["qtum"] = transfer.NewQtumTransferService()
	coinSrv["sol"] = transfer.NewSolTransferService()
	coinSrv["tlos"] = transfer.NewTlosTransferService()
	//coinSrv["pcx"] = transfer.NewPcxTransferService()
	coinSrv["ghost"] = transfer.NewGhostTransferService()
	coinSrv["dot"] = transfer.NewDotTransferService()
	coinSrv["azero"] = transfer.NewAzeroTransferService()
	coinSrv["nodle"] = transfer.NewNodleTransferService()
	coinSrv["sgb-sgb"] = transfer.NewSgbTransferService()
	coinSrv["kar"] = transfer.NewKarTransferService()
	coinSrv["bch"] = transfer.NewBchTransferService()
	coinSrv["zec"] = transfer.NewZecTransferService()
	coinSrv["dcr"] = transfer.NewDcrTransferService()
	coinSrv["btm"] = transfer.NewBtmTransferService()
	coinSrv["ckb"] = transfer.NewCkbTransferService()
	coinSrv["hc"] = transfer.NewHcTransferService()
	//coinSrv["stx"] = transfer.NewStxTransferService()
	coinSrv["stx"] = transfer.NewStxNewTransferService()
	coinSrv["nas"] = transfer.NewNasTransferService()
	coinSrv["doge"] = transfer.NewDogeTransferService()
	coinSrv["avax"] = transfer.NewAvaxTransferService()
	coinSrv["bsc"] = transfer.NewBscTransferService()
	coinSrv["fil"] = transfer.NewFilTransferService()
	coinSrv["wd"] = transfer.NewWd_wdTransferService()
	coinSrv["dash"] = transfer.NewDashTransferService()
	coinSrv["biw"] = transfer.NewBiwTransferService()
	coinSrv["atom"] = transfer.NewAtomTransferService()
	coinSrv["near"] = transfer.NewNearTransferService()
	coinSrv["yta"] = transfer.NewYtaTransferService()
	coinSrv["cfx"] = transfer.NewCfxTransferService()
	coinSrv["star"] = transfer.NewStarTransferService()
	coinSrv["fis"] = transfer.NewFisTransferService()
	coinSrv["oneo"] = transfer.NewNeoTransferService()
	coinSrv["atp"] = transfer.NewAtpTransferService()
	coinSrv["cph-cph"] = transfer.NewCphTransferService()
	coinSrv["pcx"] = transfer.NewChainXTransferService()
	//coinSrv["pcx"] = transfer.NewPcxTransferService()
	coinSrv["bcha"] = transfer.NewBchaTransferService()
	coinSrv["xec"] = transfer.NewXecTransferService()
	coinSrv["ada"] = transfer.NewAdaTransferService()
	coinSrv["trx"] = transfer.NewTrxTransferService()
	coinSrv["zen"] = transfer.NewZenTransferService()
	coinSrv["mw"] = transfer.NewMwTransferService()
	coinSrv["dip"] = transfer.NewDipTransferService()
	coinSrv["algo"] = transfer.NewAlgoTransferService()
	coinSrv["ori"] = transfer.NewOriTransferService()
	coinSrv["bos"] = transfer.NewBosTransferService()
	coinSrv["okt"] = transfer.NewOktTransferService()
	coinSrv["waves"] = transfer.NewWavesTransferService()
	coinSrv["glmr"] = transfer.NewGlmrTransferService()
	coinSrv["avaxcchain"] = transfer.NewAvaxcchainTransferService()
	coinSrv["heco"] = transfer.NewHecoTransferService()
	coinSrv["nyzo"] = transfer.NewNyzoTransferService()
	coinSrv["xdag"] = transfer.NewXdagTransferService()
	coinSrv["iost"] = transfer.NewIostTransferService()
	coinSrv["hsc"] = transfer.NewHscTransferService()
	coinSrv["dhx"] = transfer.NewDhxTransferService()
	coinSrv["dom"] = transfer.NewDomTransferService()
	coinSrv["wtc"] = transfer.NewWtcTransferService()
	coinSrv["moac"] = transfer.NewMoacTransferService()
	coinSrv["satcoin"] = transfer.NewSatcoinTransferService()
	coinSrv["eac"] = transfer.NewEacTransferService()
	coinSrv["iota"] = transfer.NewIotaTransferService()
	coinSrv["kai"] = transfer.NewKaiTransferService()
	coinSrv["rbtc"] = transfer.NewRbtcTransferService()
	coinSrv["movr"] = transfer.NewMovrTransferService()
	coinSrv["sep20"] = transfer.NewSep20TransferService()
	coinSrv["rei"] = transfer.NewReiTransferService()
	coinSrv["dscc"] = transfer.NewDsccTransferService()
	coinSrv["dscc1"] = transfer.NewDscc1TransferService()
	coinSrv["brise-brise"] = transfer.NewBriseTransferService()
	coinSrv["ccn"] = transfer.NewCcnTransferService()
	coinSrv["optim"] = transfer.NewOptimTransferService()
	coinSrv["ftm"] = transfer.NewFtmTransferService()
	coinSrv["welups"] = transfer.NewWelTransferService()
	coinSrv["rose"] = transfer.NewRoseTransferService()
	coinSrv["one"] = transfer.NewOneTransferService()
	coinSrv["rev"] = transfer.NewRevTransferService()
	coinSrv["tkm"] = transfer.NewTkmTransferService()
	coinSrv["ron"] = transfer.NewRonTransferService()
	coinSrv["neo"] = transfer.NewN3neoTransferService()
	coinSrv["flow"] = transfer.NewFlowTransferService()
	coinSrv["icp"] = transfer.NewIcpTransferService()
	coinSrv["uenc"] = transfer.NewUencTransferService()
	coinSrv["cspr"] = transfer.NewCsprTransferService()
	coinSrv["matic-matic"] = transfer.NewMaticTransferService()
	coinSrv["crust"] = transfer.NewCRustTransferService()
	coinSrv["iotx"] = transfer.NewIotexTransferService()
	coinSrv["aur"] = transfer.NewAurTransferService()
	coinSrv["evmos"] = transfer.NewEvmosTransferService()
	coinSrv["mob"] = transfer.NewMobTransferService()
	coinSrv["deso"] = transfer.NewDesoTransferService()
	coinSrv["lat"] = transfer.NewLatTransferService()
	coinSrv["hbar"] = transfer.NewHbarTransferService()
	coinSrv["steem"] = transfer.NewSteemTransferService()

	transferService = coinSrv

	//交易验证服务
	transferSecurityService = security.NewSecurityService()

	//商户订单服务
	orderService = order.NewOrderBaseService()

}

func InitDingErrBot(name, token string) {
	//钉钉通知
	ErrDingBot = &dingding.DingBot{
		Name:   name,
		Token:  token,
		Source: make(chan []byte, 50),
		Quit:   make(chan struct{}),
	}

	ErrDingBot.Start()
}
