module github.com/group-coldwallet/celo-sign

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.6.3
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9

)

replace github.com/ethereum/go-ethereum => github.com/celo-org/celo-blockchain v0.0.0-20200702195422-0af1e48c4c05
