package p2p

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"

	cmn "btmSign/bytom/lib/github.com/tendermint/tmlibs/common"
	"btmSign/bytom/lib/github.com/tendermint/tmlibs/flowrate"
	"github.com/btcsuite/go-socks/socks"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	cfg "btmSign/bytom/config"
	"btmSign/bytom/consensus"
	"btmSign/bytom/crypto/ed25519/chainkd"
	"btmSign/bytom/p2p/connection"
)

// peerConn contains the raw connection and its config.
type peerConn struct {
	outbound bool
	config   *PeerConfig
	conn     net.Conn // source connection
}

// PeerConfig is a Peer configuration.
type PeerConfig struct {
	HandshakeTimeout time.Duration           `mapstructure:"handshake_timeout"` // times are in seconds
	DialTimeout      time.Duration           `mapstructure:"dial_timeout"`
	ProxyAddress     string                  `mapstructure:"proxy_address"`
	ProxyUsername    string                  `mapstructure:"proxy_username"`
	ProxyPassword    string                  `mapstructure:"proxy_password"`
	MConfig          *connection.MConnConfig `mapstructure:"connection"`
}

// DefaultPeerConfig returns the default config.
func DefaultPeerConfig(config *cfg.P2PConfig) *PeerConfig {
	return &PeerConfig{
		HandshakeTimeout: time.Duration(config.HandshakeTimeout) * time.Second, // * time.Second,
		DialTimeout:      time.Duration(config.DialTimeout) * time.Second,      // * time.Second,
		ProxyAddress:     config.ProxyAddress,
		ProxyUsername:    config.ProxyUsername,
		ProxyPassword:    config.ProxyPassword,
		MConfig:          connection.DefaultMConnConfig(),
	}
}

// Peer represent a bytom network node
type Peer struct {
	cmn.BaseService
	*NodeInfo
	*peerConn
	mconn *connection.MConnection // multiplex connection
	Key   string
	isLAN bool
}

func (p *Peer) Moniker() string {
	return p.NodeInfo.Moniker
}

// OnStart implements BaseService.
func (p *Peer) OnStart() error {
	p.BaseService.OnStart()
	//return p.mconn.Start()
	return nil
}

// OnStop implements BaseService.
func (p *Peer) OnStop() {
	p.BaseService.OnStop()
	p.mconn.Stop()
}

func newPeer(pc *peerConn, nodeInfo *NodeInfo, reactorsByCh map[byte]Reactor, chDescs []*connection.ChannelDescriptor, onPeerError func(*Peer, interface{}), isLAN bool) *Peer {
	// Key and NodeInfo are set after Handshake
	p := &Peer{
		peerConn: pc,
		NodeInfo: nodeInfo,
		Key:      hex.EncodeToString(nodeInfo.PubKey),
		isLAN:    isLAN,
	}
	p.mconn = createMConnection(pc.conn, p, reactorsByCh, chDescs, onPeerError, pc.config.MConfig)
	p.BaseService = *cmn.NewBaseService(nil, "Peer", p)
	return p
}

func newOutboundPeerConn(addr *NetAddress, ourNodePrivKey chainkd.XPrv, config *PeerConfig) (*peerConn, error) {
	conn, err := dial(addr, config)
	if err != nil {
		return nil, errors.Wrap(err, "Error dial peer")
	}

	pc, err := newPeerConn(conn, true, ourNodePrivKey, config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return pc, nil
}

func newInboundPeerConn(conn net.Conn, ourNodePrivKey chainkd.XPrv, config *cfg.P2PConfig) (*peerConn, error) {
	return newPeerConn(conn, false, ourNodePrivKey, DefaultPeerConfig(config))
}

func newPeerConn(rawConn net.Conn, outbound bool, ourNodePrivKey chainkd.XPrv, config *PeerConfig) (*peerConn, error) {
	rawConn.SetDeadline(time.Now().Add(config.HandshakeTimeout))
	conn, err := connection.MakeSecretConnection(rawConn, ourNodePrivKey)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating peer")
	}

	return &peerConn{
		config:   config,
		outbound: outbound,
		conn:     conn,
	}, nil
}

// Addr returns peer's remote network address.
func (p *Peer) Addr() net.Addr {
	return p.conn.RemoteAddr()
}

// CanSend returns true if the send queue is not full, false otherwise.
func (p *Peer) CanSend(chID byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.CanSend(chID)
}

// CloseConn should be used when the peer was created, but never started.
func (pc *peerConn) CloseConn() {
	pc.conn.Close()
}

// Equals reports whenever 2 peers are actually represent the same node.
func (p *Peer) Equals(other *Peer) bool {
	return p.Key == other.Key
}

// HandshakeTimeout performs a handshake between a given node and the peer.
// NOTE: blocking
func (pc *peerConn) HandshakeTimeout(ourNodeInfo *NodeInfo, timeout time.Duration) (*NodeInfo, error) {
	// Set deadline for handshake so we don't block forever on conn.ReadFull
	if err := pc.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	var peerNodeInfo = new(NodeInfo)
	//writeTask := func(i int) (val interface{}, err error, about bool) {
	//	var n int
	//	wire.WriteBinary(ourNodeInfo, pc.conn, &n, &err)
	//	return nil, err, false
	//}
	//
	//readTask := func(i int) (val interface{}, err error, about bool) {
	//	var n int
	//	wire.ReadBinary(peerNodeInfo, pc.conn, maxNodeInfoSize, &n, &err)
	//	return nil, err, false
	//
	//}
	//cmn.Parallel(writeTask, readTask)

	// In parallel, handle reads and writes
	//trs, ok := cmn.Parallel(writeTask, readTask)
	//if !ok {
	//	return nil, errors.New("Parallel task run failed")
	//}
	//for i := 0; i < 2; i++ {
	//	res, ok := trs.LatestResult(i)
	//	if !ok {
	//		return nil, fmt.Errorf("Task %d did not complete", i)
	//	}
	//
	//	if res.Error != nil {
	//		return nil, errors.Wrap(res.Error, fmt.Sprintf("Task %d got error", i))
	//	}
	//}

	// Remove deadline
	if err := pc.conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}
	peerNodeInfo.RemoteAddr = pc.conn.RemoteAddr().String()
	return peerNodeInfo, nil
}

// ID return the uuid of the peer
func (p *Peer) ID() string {
	return p.Key
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *Peer) IsOutbound() bool {
	return p.outbound
}

// IsLAN returns true if peer is LAN peer, false otherwise.
func (p *Peer) IsLAN() bool {
	return p.isLAN
}

// PubKey returns peer's public key.
func (p *Peer) PubKey() ed25519.PublicKey {
	return p.conn.(*connection.SecretConnection).RemotePubKey()
}

// Send msg to the channel identified by chID byte. Returns false if the send
// queue is full after timeout, specified by MConnection.
func (p *Peer) Send(chID byte, msg interface{}) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.Send(chID, msg)
}

// ServiceFlag return the ServiceFlag of this peer
func (p *Peer) ServiceFlag() consensus.ServiceFlag {
	services := consensus.SFFullNode
	if len(p.Other) == 0 {
		return services
	}

	if serviceFlag, err := strconv.ParseUint(p.Other[0], 10, 64); err == nil {
		services = consensus.ServiceFlag(serviceFlag)
	}
	return services
}

// String representation.
func (p *Peer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v %v out}", p.mconn, p.Key[:12])
	}
	return fmt.Sprintf("Peer{%v %v in}", p.mconn, p.Key[:12])
}

// TrafficStatus return the in and out traffic status
func (p *Peer) TrafficStatus() (*flowrate.Status, *flowrate.Status) {
	return p.mconn.TrafficStatus()
}

// TrySend msg to the channel identified by chID byte. Immediately returns
// false if the send queue is full.
func (p *Peer) TrySend(chID byte, msg interface{}) bool {
	if !p.IsRunning() {
		return false
	}

	log.WithFields(log.Fields{
		"module": logModule,
		"peer":   p.Addr(),
		"msg":    msg,
		"type":   reflect.TypeOf(msg),
	}).Info("send message to peer")
	return p.mconn.TrySend(chID, msg)
}

func createMConnection(conn net.Conn, p *Peer, reactorsByCh map[byte]Reactor, chDescs []*connection.ChannelDescriptor, onPeerError func(*Peer, interface{}), config *connection.MConnConfig) *connection.MConnection {
	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			cmn.PanicSanity(cmn.Fmt("Unknown channel %X", chID))
		}
		reactor.Receive(chID, p, msgBytes)
	}

	onError := func(r interface{}) {
		onPeerError(p, r)
	}
	return connection.NewMConnectionWithConfig(conn, chDescs, onReceive, onError, config)
}

func dial(addr *NetAddress, config *PeerConfig) (net.Conn, error) {
	var conn net.Conn
	var err error
	if config.ProxyAddress == "" {
		conn, err = addr.DialTimeout(config.DialTimeout)
	} else {
		proxy := &socks.Proxy{
			Addr:         config.ProxyAddress,
			Username:     config.ProxyUsername,
			Password:     config.ProxyPassword,
			TorIsolation: false,
		}
		conn, err = addr.DialTimeoutWithProxy(proxy, config.DialTimeout)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}
