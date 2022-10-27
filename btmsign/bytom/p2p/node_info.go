package p2p

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	cfg "btmSign/bytom/config"
	"btmSign/bytom/consensus"
	"btmSign/bytom/version"
)

const maxNodeInfoSize = 10240 // 10Kb

// NodeInfo peer node info
type NodeInfo struct {
	PubKey     []byte `json:"pub_key"`
	Moniker    string `json:"moniker"`
	Network    string `json:"network"`
	RemoteAddr string `json:"remote_addr"`
	ListenAddr string `json:"listen_addr"`
	Version    string `json:"version"` // major.minor.revision
	// other application specific data
	// field 0: node service flags. field 1: node alias.
	Other []string `json:"other"`
}

func NewNodeInfo(config *cfg.Config, pubkey ed25519.PublicKey, listenAddr string) *NodeInfo {
	other := []string{strconv.FormatUint(uint64(consensus.DefaultServices), 10)}
	if config.NodeAlias != "" {
		other = append(other, config.NodeAlias)
	}
	return &NodeInfo{
		PubKey:     pubkey,
		Moniker:    config.Moniker,
		Network:    config.ChainID,
		ListenAddr: listenAddr,
		Version:    version.Version,
		Other:      other,
	}
}

// CompatibleWith checks if two NodeInfo are compatible with eachother.
// CONTRACT: two nodes are compatible if the major version matches and network match
func (info *NodeInfo) CompatibleWith(other *NodeInfo) error {
	compatible, err := version.CompatibleWith(other.Version)
	if err != nil {
		return err
	}
	if !compatible {
		return fmt.Errorf("Peer is on a different major version. Peer version: %v, node version: %v", other.Version, info.Version)
	}

	if info.Network != other.Network {
		return fmt.Errorf("Peer is on a different network. Peer network: %v, node network: %v", other.Network, info.Network)
	}
	return nil
}

func (info NodeInfo) DoFilter(ip string, pubKey string) error {
	if hex.EncodeToString(info.PubKey) == pubKey {
		return ErrConnectSelf
	}

	return nil
}

// ListenHost peer listener ip address
func (info *NodeInfo) listenHost() string {
	host, _, _ := net.SplitHostPort(info.ListenAddr)
	return host
}

// RemoteAddrHost peer external ip address
func (info *NodeInfo) RemoteAddrHost() string {
	host, _, _ := net.SplitHostPort(info.RemoteAddr)
	return host
}

// GetNetwork get node info network field
func (info *NodeInfo) GetNetwork() string {
	return info.Network
}

// String representation
func (info NodeInfo) String() string {
	return fmt.Sprintf("NodeInfo{pk: %v, moniker: %v, network: %v [listen %v], version: %v (%v)}", info.PubKey, info.Moniker, info.Network, info.ListenAddr, info.Version, info.Other)
}
