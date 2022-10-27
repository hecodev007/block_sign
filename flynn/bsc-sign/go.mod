module github.com/group-coldwallet/bsc-sign

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.6.3
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.7.0
)

replace github.com/ethereum/go-ethereum => github.com/binance-chain/bsc v1.0.4
