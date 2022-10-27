package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"btmSign/bytom/contract"
	cmn "btmSign/bytom/lib/github.com/tendermint/tmlibs/common"
	"github.com/kr/secureheader"
	log "github.com/sirupsen/logrus"

	"btmSign/bytom/accesstoken"
	cfg "btmSign/bytom/config"
	"btmSign/bytom/dashboard/equity"
	"btmSign/bytom/errors"
	"btmSign/bytom/event"
	"btmSign/bytom/net/http/authn"
	"btmSign/bytom/net/http/gzip"
	"btmSign/bytom/net/http/httpjson"
	"btmSign/bytom/net/http/static"
	"btmSign/bytom/net/websocket"
	"btmSign/bytom/netsync/peers"
	"btmSign/bytom/p2p"
	"btmSign/bytom/proposal/blockproposer"
	"btmSign/bytom/protocol"
	"btmSign/bytom/wallet"
)

var (
	errNotAuthenticated = errors.New("not authenticated")
	httpReadTimeout     = 2 * time.Minute
	httpWriteTimeout    = time.Hour
)

const (
	// SUCCESS indicates the rpc calling is successful.
	SUCCESS = "success"
	// FAIL indicated the rpc calling is failed.
	FAIL      = "fail"
	logModule = "api"
)

// Response describes the response standard.
type Response struct {
	Status      string      `json:"status,omitempty"`
	Code        string      `json:"code,omitempty"`
	Msg         string      `json:"msg,omitempty"`
	ErrorDetail string      `json:"error_detail,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

// NewSuccessResponse success response
func NewSuccessResponse(data interface{}) Response {
	return Response{Status: SUCCESS, Data: data}
}

// FormatErrResp format error response
func FormatErrResp(err error) (response Response) {
	response = Response{Status: FAIL}
	root := errors.Root(err)
	// Some types cannot be used as map keys, for example slices.
	// If an error's underlying type is one of these, don't panic.
	// Just treat it like any other missing entry.
	defer func() {
		if err := recover(); err != nil {
			response.ErrorDetail = ""
		}
	}()

	if info, ok := respErrFormatter[root]; ok {
		response.Code = info.ChainCode
		response.Msg = info.Message
		response.ErrorDetail = err.Error()
	} else {
		response.Code = respErrFormatter[ErrDefault].ChainCode
		response.Msg = respErrFormatter[ErrDefault].Message
		response.ErrorDetail = err.Error()
	}
	return response
}

// NewErrorResponse error response
func NewErrorResponse(err error) Response {
	response := FormatErrResp(err)
	return response
}

type waitHandler struct {
	h  http.Handler
	wg sync.WaitGroup
}

func (wh *waitHandler) Set(h http.Handler) {
	wh.h = h
	wh.wg.Done()
}

func (wh *waitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	wh.wg.Wait()
	wh.h.ServeHTTP(w, req)
}

// API is the scheduling center for server
type API struct {
	sync            NetSync
	wallet          *wallet.Wallet
	accessTokens    *accesstoken.CredentialStore
	chain           *protocol.Chain
	contractTracer  *contract.TraceService
	server          *http.Server
	handler         http.Handler
	blockProposer   *blockproposer.BlockProposer
	notificationMgr *websocket.WSNotificationManager
	eventDispatcher *event.Dispatcher
}

func (a *API) initServer(config *cfg.Config) {
	// The waitHandler accepts incoming requests, but blocks until its underlying
	// handler is set, when the second phase is complete.
	var coreHandler waitHandler
	var handler http.Handler

	coreHandler.wg.Add(1)
	mux := http.NewServeMux()
	mux.Handle("/", &coreHandler)

	handler = AuthHandler(mux, a.accessTokens, config.Auth.Disable)
	handler = RedirectHandler(handler)

	secureheader.DefaultConfig.PermitClearLoopback = true
	secureheader.DefaultConfig.HTTPSRedirect = false
	secureheader.DefaultConfig.Next = handler

	a.server = &http.Server{
		// Note: we should not set TLSConfig here;
		// we took care of TLS with the listener in maybeUseTLS.
		Handler:      secureheader.DefaultConfig,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		// Disable HTTP/2 for now until the Go implementation is more stable.
		// https://github.com/golang/go/issues/16450
		// https://github.com/golang/go/issues/17071
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}

	coreHandler.Set(a)
}

// StartServer start the server
func (a *API) StartServer(address string) {
	log.WithFields(log.Fields{"module": logModule, "api address:": address}).Info("Rpc listen")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		cmn.Exit(cmn.Fmt("Failed to register tcp port: %v", err))
	}

	// The `Serve` call has to happen in its own goroutine because
	// it's blocking and we need to proceed to the rest of the core setup after
	// we call it.
	go func() {
		if err := a.server.Serve(listener); err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": errors.Wrap(err, "Serve")}).Error("Rpc server")
		}
	}()
}

type NetSync interface {
	IsListening() bool
	IsCaughtUp() bool
	PeerCount() int
	GetNetwork() string
	BestPeer() *peers.PeerInfo
	DialPeerWithAddress(addr *p2p.NetAddress) error
	GetPeerInfos() []*peers.PeerInfo
	StopPeer(peerID string) error
}

// NewAPI create and initialize the API
func NewAPI(sync NetSync, wallet *wallet.Wallet, blockProposer *blockproposer.BlockProposer, chain *protocol.Chain, traceService *contract.TraceService, config *cfg.Config, token *accesstoken.CredentialStore, dispatcher *event.Dispatcher, notificationMgr *websocket.WSNotificationManager) *API {
	api := &API{
		sync:            sync,
		wallet:          wallet,
		chain:           chain,
		contractTracer:  traceService,
		accessTokens:    token,
		blockProposer:   blockProposer,
		eventDispatcher: dispatcher,
		notificationMgr: notificationMgr,
	}
	api.buildHandler()
	api.initServer(config)

	return api
}

func (a *API) SetWallet(wallets *wallet.Wallet) {
	a.wallet = wallets
}

func (a *API) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.handler.ServeHTTP(rw, req)
}

// buildHandler is in charge of all the rpc handling.
func (a *API) buildHandler() {
	walletEnable := false
	m := http.NewServeMux()

	if a.wallet != nil {
		walletEnable = true
		m.Handle("/create-account", jsonHandler(a.createAccount))
		m.Handle("/update-account-alias", jsonHandler(a.updateAccountAlias))
		m.Handle("/list-accounts", jsonHandler(a.listAccounts))
		m.Handle("/delete-account", jsonHandler(a.deleteAccount))

		m.Handle("/create-account-receiver", jsonHandler(a.createAccountReceiver))
		m.Handle("/list-addresses", jsonHandler(a.listAddresses))
		m.Handle("/validate-address", jsonHandler(a.validateAddress))
		m.Handle("/list-pubkeys", jsonHandler(a.listPubKeys))

		m.Handle("/get-mining-address", jsonHandler(a.getMiningAddress))
		m.Handle("/set-mining-address", jsonHandler(a.setMiningAddress))

		m.Handle("/create-asset", jsonHandler(a.createAsset))
		m.Handle("/update-asset-alias", jsonHandler(a.updateAssetAlias))
		m.Handle("/get-asset", jsonHandler(a.getAsset))
		m.Handle("/list-assets", jsonHandler(a.listAssets))

		m.Handle("/create-key", jsonHandler(a.pseudohsmCreateKey))
		m.Handle("/update-key-alias", jsonHandler(a.pseudohsmUpdateKeyAlias))
		m.Handle("/list-keys", jsonHandler(a.pseudohsmListKeys))
		m.Handle("/delete-key", jsonHandler(a.pseudohsmDeleteKey))
		m.Handle("/reset-key-password", jsonHandler(a.pseudohsmResetPassword))
		m.Handle("/check-key-password", jsonHandler(a.pseudohsmCheckPassword))
		m.Handle("/sign-message", jsonHandler(a.signMessage))

		m.Handle("/build-transaction", jsonHandler(a.Build))
		m.Handle("/build-chain-transactions", jsonHandler(a.buildChainTxs))
		m.Handle("/sign-transaction", jsonHandler(a.signTemplate))
		m.Handle("/sign-transactions", jsonHandler(a.signTemplates))

		m.Handle("/get-transaction", jsonHandler(a.getTransaction))
		m.Handle("/list-transactions", jsonHandler(a.listTransactions))

		m.Handle("/list-balances", jsonHandler(a.listBalances))
		m.Handle("/list-unspent-outputs", jsonHandler(a.listUnspentOutputs))
		m.Handle("/list-account-votes", jsonHandler(a.listAccountVotes))

		m.Handle("/decode-program", jsonHandler(a.decodeProgram))

		m.Handle("/backup-wallet", jsonHandler(a.backupWalletImage))
		m.Handle("/restore-wallet", jsonHandler(a.restoreWalletImage))
		m.Handle("/rescan-wallet", jsonHandler(a.rescanWallet))
		m.Handle("/wallet-info", jsonHandler(a.getWalletInfo))
		m.Handle("/recovery-wallet", jsonHandler(a.recoveryFromRootXPubs))
	} else {
		log.Warn("Please enable wallet")
	}

	m.Handle("/", alwaysError(errors.New("not Found")))
	m.Handle("/error", jsonHandler(a.walletError))

	m.Handle("/create-access-token", jsonHandler(a.createAccessToken))
	m.Handle("/list-access-tokens", jsonHandler(a.listAccessTokens))
	m.Handle("/delete-access-token", jsonHandler(a.deleteAccessToken))
	m.Handle("/check-access-token", jsonHandler(a.checkAccessToken))

	m.Handle("/create-contract", jsonHandler(a.createContract))
	m.Handle("/update-contract-alias", jsonHandler(a.updateContractAlias))
	m.Handle("/get-contract", jsonHandler(a.getContract))
	m.Handle("/list-contracts", jsonHandler(a.listContracts))

	m.Handle("/submit-transaction", jsonHandler(a.submit))
	m.Handle("/submit-transactions", jsonHandler(a.submitTxs))
	m.Handle("/estimate-transaction-gas", jsonHandler(a.estimateTxGas))
	m.Handle("/estimate-chain-transaction-gas", jsonHandler(a.estimateChainTxGas))

	m.Handle("/get-unconfirmed-transaction", jsonHandler(a.getUnconfirmedTx))
	m.Handle("/list-unconfirmed-transactions", jsonHandler(a.listUnconfirmedTxs))
	m.Handle("/decode-raw-transaction", jsonHandler(a.decodeRawTransaction))

	m.Handle("/get-block", jsonHandler(a.getBlock))
	m.Handle("/get-raw-block", jsonHandler(a.getRawBlock))
	m.Handle("/get-block-hash", jsonHandler(a.getBestBlockHash))
	m.Handle("/get-block-header", jsonHandler(a.getBlockHeader))
	m.Handle("/get-block-count", jsonHandler(a.getBlockCount))

	m.Handle("/is-mining", jsonHandler(a.isMining))
	m.Handle("/set-mining", jsonHandler(a.setMining))

	m.Handle("/verify-message", jsonHandler(a.verifyMessage))

	m.Handle("/gas-rate", jsonHandler(a.gasRate))
	m.Handle("/net-info", jsonHandler(a.getNetInfo))
	m.Handle("/chain-status", jsonHandler(a.getChainStatus))

	m.Handle("/list-peers", jsonHandler(a.listPeers))
	m.Handle("/disconnect-peer", jsonHandler(a.disconnectPeer))
	m.Handle("/connect-peer", jsonHandler(a.connectPeer))

	m.Handle("/get-merkle-proof", jsonHandler(a.getMerkleProof))
	m.Handle("/get-vote-result", jsonHandler(a.getVoteResult))

	m.Handle("/get-contract-instance", jsonHandler(a.getContractInstance))
	m.Handle("/create-contract-instance", jsonHandler(a.createContractInstance))
	m.Handle("/remove-contract-instance", jsonHandler(a.removeContractInstance))

	m.HandleFunc("/websocket-subscribe", a.websocketHandler)

	handler := walletHandler(m, walletEnable)
	handler = webAssetsHandler(handler)
	handler = gzip.Handler{Handler: handler}
	a.handler = handler
}

// json Handler
func jsonHandler(f interface{}) http.Handler {
	h, err := httpjson.Handler(f, errorFormatter.Write)
	if err != nil {
		panic(err)
	}
	return h
}

// error Handler
func alwaysError(err error) http.Handler {
	return jsonHandler(func() error { return err })
}

func webAssetsHandler(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/dashboard/", http.StripPrefix("/dashboard/", static.Handler{
		//Assets:  dashboard.Files,
		Assets:  nil,
		Default: "index.html",
	}))
	mux.Handle("/equity/", http.StripPrefix("/equity/", static.Handler{
		Assets:  equity.Files,
		Default: "index.html",
	}))
	mux.Handle("/", next)

	return mux
}

// AuthHandler access token auth Handler
func AuthHandler(handler http.Handler, accessTokens *accesstoken.CredentialStore, authDisable bool) http.Handler {
	authenticator := authn.NewAPI(accessTokens, authDisable)

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// TODO(tessr): check that this path exists; return early if this path isn't legit
		req, err := authenticator.Authenticate(req)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": errors.Wrap(err, "Serve")}).Error("Authenticate fail")
			err = errors.WithDetail(errNotAuthenticated, err.Error())
			errorFormatter.Write(req.Context(), rw, err)
			return
		}
		handler.ServeHTTP(rw, req)
	})
}

// RedirectHandler redirect to dashboard handler
func RedirectHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/dashboard/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func walletHandler(m *http.ServeMux, walletEnable bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// when the wallet is not been opened and the url path is not been found, modify url path to error,
		// and redirect handler to error
		if _, pattern := m.Handler(req); pattern != req.URL.Path && !walletEnable {
			req.URL.Path = "/error"
			walletRedirectHandler(w, req)
			return
		}

		m.ServeHTTP(w, req)
	})
}

// walletRedirectHandler redirect to error when the wallet is closed
func walletRedirectHandler(w http.ResponseWriter, req *http.Request) {
	h := http.RedirectHandler(req.URL.String(), http.StatusMovedPermanently)
	h.ServeHTTP(w, req)
}
