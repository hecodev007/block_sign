package node

import (
	"net"
	"net/http"
	_ "net/http/pprof"

	cmn "btmSign/bytom/lib/github.com/tendermint/tmlibs/common"
	log "github.com/sirupsen/logrus"
	//browser "github.com/toqueteos/webbrowser"

	"btmSign/bytom/proposal/blockproposer"
	//"github.com/prometheus/prometheus/util/flock"

	"btmSign/bytom/accesstoken"
	"btmSign/bytom/account"
	"btmSign/bytom/api"
	"btmSign/bytom/asset"
	"btmSign/bytom/blockchain/pseudohsm"
	cfg "btmSign/bytom/config"
	"btmSign/bytom/consensus"
	"btmSign/bytom/contract"
	"btmSign/bytom/database"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/env"
	"btmSign/bytom/event"
	bytomLog "btmSign/bytom/log"
	"btmSign/bytom/net/websocket"
	"btmSign/bytom/netsync"
	"btmSign/bytom/protocol"
	w "btmSign/bytom/wallet"
)

const (
	webHost   = "http://127.0.0.1"
	logModule = "node"
)

// Node represent bytom node
type Node struct {
	cmn.BaseService

	config          *cfg.Config
	eventDispatcher *event.Dispatcher
	syncManager     *netsync.SyncManager

	wallet          *w.Wallet
	accessTokens    *accesstoken.CredentialStore
	notificationMgr *websocket.WSNotificationManager
	api             *api.API
	chain           *protocol.Chain
	traceService    *contract.TraceService
	blockProposer   *blockproposer.BlockProposer
	miningEnable    bool
}

// NewNode create bytom node
func NewNode(config *cfg.Config) *Node {
	if err := initNodeConfig(config); err != nil {
		cmn.Exit(cmn.Fmt("Failed to init config: %v", err))
	}

	// Get store
	if config.DBBackend != "memdb" && config.DBBackend != "leveldb" {
		cmn.Exit(cmn.Fmt("Param db_backend [%v] is invalid, use leveldb or memdb", config.DBBackend))
	}
	coreDB := dbm.NewDB("core", config.DBBackend, config.DBDir())
	store := database.NewStore(coreDB)

	tokenDB := dbm.NewDB("accesstoken", config.DBBackend, config.DBDir())
	accessTokens := accesstoken.NewStore(tokenDB)

	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)

	chain, err := protocol.NewChain(store, txPool, dispatcher)
	if err != nil {
		cmn.Exit(cmn.Fmt("Failed to create chain structure: %v", err))
	}

	traceService := startTraceUpdater(chain, config)

	var accounts *account.Manager
	var assets *asset.Registry
	var wallet *w.Wallet

	hsm, err := pseudohsm.New(config.KeysDir())
	if err != nil {
		cmn.Exit(cmn.Fmt("initialize HSM failed: %v", err))
	}

	if !config.Wallet.Disable {
		walletDB := dbm.NewDB("wallet", config.DBBackend, config.DBDir())
		accounts = account.NewManager(walletDB, chain)
		assets = asset.NewRegistry(walletDB, chain)
		contracts := contract.NewRegistry(walletDB)
		wallet, err = w.NewWallet(walletDB, accounts, assets, contracts, hsm, chain, dispatcher, config.Wallet.TxIndex)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": err}).Error("init NewWallet")
		}

		// trigger rescan wallet
		if config.Wallet.Rescan {
			wallet.RescanBlocks()
		}
	}

	fastSyncDB := dbm.NewDB("fastsync", config.DBBackend, config.DBDir())
	syncManager, err := netsync.NewSyncManager(config, chain, txPool, dispatcher, fastSyncDB)
	if err != nil {
		cmn.Exit(cmn.Fmt("Failed to create sync manager: %v", err))
	}

	notificationMgr := websocket.NewWsNotificationManager(config.Websocket.MaxNumWebsockets, config.Websocket.MaxNumConcurrentReqs, chain, dispatcher)

	// run the profile server
	profileHost := config.ProfListenAddress
	if profileHost != "" {
		// Profiling bytomd programs.see (https://blog.golang.org/profiling-go-programs)
		// go tool pprof http://profileHose/debug/pprof/heap
		go func() {
			if err = http.ListenAndServe(profileHost, nil); err != nil {
				cmn.Exit(cmn.Fmt("Failed to register tcp profileHost: %v", err))
			}
		}()
	}

	node := &Node{
		eventDispatcher: dispatcher,
		config:          config,
		syncManager:     syncManager,
		accessTokens:    accessTokens,
		wallet:          wallet,
		chain:           chain,
		traceService:    traceService,
		miningEnable:    config.Mining,
		notificationMgr: notificationMgr,
	}

	node.BaseService = *cmn.NewBaseService(nil, "Node", node)
	node.blockProposer = blockproposer.NewBlockProposer(chain, accounts, dispatcher)
	return node
}

func startTraceUpdater(chain *protocol.Chain, cfg *cfg.Config) *contract.TraceService {
	db := dbm.NewDB("trace", cfg.DBBackend, cfg.DBDir())
	store := contract.NewTraceStore(db)
	tracerService := contract.NewTraceService(contract.NewInfrastructure(chain, store))
	traceUpdater := contract.NewTraceUpdater(tracerService, chain)
	go traceUpdater.Sync()
	return tracerService
}

func initNodeConfig(config *cfg.Config) error {
	if err := lockDataDirectory(config); err != nil {
		cmn.Exit("Error: " + err.Error())
	}

	if err := bytomLog.InitLogFile(config); err != nil {
		log.WithField("err", err).Fatalln("InitLogFile failed")
	}

	initActiveNetParams(config)
	initCommonConfig(config)
	return nil
}

// Lock data directory after daemonization
func lockDataDirectory(config *cfg.Config) error {
	//_, _, err := flock.New(filepath.Join(config.RootDir, "LOCK"))
	//if err != nil {
	//	return errors.New("datadir already used by another process")
	//}
	return nil
}

func initActiveNetParams(config *cfg.Config) {
	var exist bool
	consensus.ActiveNetParams, exist = consensus.NetParams[config.ChainID]
	if !exist {
		cmn.Exit(cmn.Fmt("chain_id[%v] don't exist", config.ChainID))
	}
}

func initCommonConfig(config *cfg.Config) {
	cfg.CommonConfig = config
}

// Lanch web broser or not
func launchWebBrowser(port string) {
	//webAddress := webHost + ":" + port
	//log.Info("Launching System Browser with :", webAddress)
	//if err := browser.Open(webAddress); err != nil {
	//	log.Error(err.Error())
	//	return
	//}
	return
}

func (n *Node) initAndstartAPIServer() {
	n.api = api.NewAPI(n.syncManager, n.wallet, n.blockProposer, n.chain, n.traceService, n.config, n.accessTokens, n.eventDispatcher, n.notificationMgr)

	listenAddr := env.String("LISTEN", n.config.ApiAddress)
	env.Parse()
	n.api.StartServer(*listenAddr)
}

func (n *Node) OnStart() error {
	if n.miningEnable {
		if _, err := n.wallet.AccountMgr.GetMiningAddress(); err != nil {
			n.miningEnable = false
			log.Error(err)
		} else {
			n.blockProposer.Start()
		}
	}
	if !n.config.VaultMode {
		if err := n.syncManager.Start(); err != nil {
			return err
		}
	}

	n.initAndstartAPIServer()
	if err := n.notificationMgr.Start(); err != nil {
		return err
	}

	if !n.config.Web.Closed {
		_, port, err := net.SplitHostPort(n.config.ApiAddress)
		if err != nil {
			log.Error("Invalid api address")
			return err
		}
		launchWebBrowser(port)
	}
	return nil
}

func (n *Node) OnStop() {
	n.notificationMgr.Shutdown()
	n.notificationMgr.WaitForShutdown()
	n.BaseService.OnStop()
	if n.miningEnable {
		n.blockProposer.Stop()
	}
	if !n.config.VaultMode {
		n.syncManager.Stop()
	}
	n.eventDispatcher.Stop()
}

func (n *Node) RunForever() {
	// Sleep forever and then...
	cmn.TrapSignal(func() {
		n.Stop()
	})
}
