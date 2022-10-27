module marSign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/eoscanada/eos-go v0.9.0
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gin-gonic/gin v1.4.0
	github.com/ilius/crock32 v0.0.0-20200913102936-af44b4eacbe6
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/okex/okexchain-go-sdk v0.16.0
	github.com/shopspring/decimal v1.2.0
	github.com/tidwall/gjson v1.3.2
	github.com/ugorji/go v1.1.7 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb
	google.golang.org/appengine v1.6.1
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/okex/cosmos-sdk v0.39.2-okexchain2
	github.com/tendermint/tendermint => github.com/okex/tendermint v0.33.9-okexchain1
)
