module arpsign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.39.2
	github.com/eoscanada/eos-go v0.9.0
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gin-gonic/gin v1.4.0
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/okex/exchain v0.18.2
	github.com/okex/exchain-go-sdk v0.18.0
	github.com/onethefour/common v0.0.4
	github.com/shopspring/decimal v1.2.0
	github.com/ugorji/go v1.1.7 // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb
	google.golang.org/appengine v1.6.1
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/okex/cosmos-sdk v0.39.2-exchain3
	github.com/tendermint/iavl => github.com/okex/iavl v0.14.3-exchain
	github.com/tendermint/tendermint => github.com/okex/tendermint v0.33.9-exchain2
)
