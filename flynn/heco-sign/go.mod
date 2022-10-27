module github.com/group-coldwallet/heco-sign

go 1.15

replace github.com/ethereum/go-ethereum => github.com/HuobiGroup/huobi-eco-chain v1.0.0

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ethereum/go-ethereum v1.10.14
	github.com/gin-gonic/gin v1.6.3
	github.com/prometheus/common v0.0.0-20181113130724-41aa239b4cce
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
