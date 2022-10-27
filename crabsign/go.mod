module wallet-sign

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ChainSafe/go-schnorrkel v0.0.0-20210318173838-ccb5cd955283
	github.com/ChainSafe/gossamer v0.6.0
	github.com/Dipper-Labs/Dipper-Protocol v0.0.0-20201103114409-9e306c5ed78f
	github.com/Dipper-Labs/go-sdk v1.0.3
	github.com/JFJun/arweave-go v0.0.0-20200525082925-be2aa616e219
	github.com/JFJun/bifrost-go v1.0.3
	github.com/JFJun/go-substrate-crypto v1.0.1
	github.com/JFJun/helium-go v1.0.0
	github.com/JFJun/near-go v0.0.3
	github.com/JFJun/rpc-tool v0.0.0-20200417031616-ffc1cf0665bf
	github.com/JFJun/solana-go v1.0.2
	github.com/JFJun/trx-sign-go v1.0.3
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/coldwallet-group/bifrost-go v1.1.2
	github.com/coldwallet-group/stafi-substrate-go v1.2.0
	github.com/coldwallet-group/substrate-go v1.3.2
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/ethereum/go-ethereum v1.10.12
	github.com/fbsobreira/gotron-sdk v0.0.0-20201030191254-389aec83c8f9
	github.com/fioprotocol/fio-go v1.0.0-alpha3
	github.com/gin-gonic/gin v1.6.2
	github.com/go-redis/redis/v8 v8.11.4
	github.com/goat-systems/go-tezos/v4 v4.0.4
	github.com/gorilla/websocket v1.4.2
	github.com/itering/subscan v0.1.0
	github.com/itering/substrate-api-rpc v0.4.11
	github.com/mendsley/gojwk v0.0.0-20141217222730-4d5ec6e58103
	github.com/onethefour/common v0.0.4
	github.com/pierrec/xxHash v0.1.5
	github.com/prometheus/common v0.9.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rs/cors v1.8.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.5.0
	github.com/stafiprotocol/go-substrate-rpc-client v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.32.13
	github.com/vedhavyas/go-subkey v1.0.2
	github.com/yanyushr/go-substrate-rpc-client/v3 v3.2.1
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/yaml.v2 v2.4.0
	gxclient-go v0.0.0-00010101000000-000000000000
)

replace github.com/ChainSafe/go-schnorrkel v1.0.0 => github.com/ChainSafe/go-schnorrkel v0.0.0-20210127175223-0f934d64ecac

replace gxclient-go => github.com/gxchain/gxclient-go v0.0.0-20200407072930-278858fef96e
