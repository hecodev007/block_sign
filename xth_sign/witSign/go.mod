module witSign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/btcsuite/golangcrypto v0.0.0-20150304025918-53f62d9b43e8
	github.com/eoscanada/eos-go v0.9.0
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/okex/okchain-go-sdk v0.11.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/tebeka/strftime v0.1.3 // indirect
	github.com/tendermint/tendermint v0.33.9 // indirect
	github.com/ugorji/go v1.1.7 // indirect
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb // indirect
	golang.org/x/sys v0.0.0-20201018230417-eeed37f84f13 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/okex/cosmos-sdk v0.37.9-okexchain8
	github.com/tendermint/iavl => github.com/okex/iavl v0.12.4-okexchain
	github.com/tendermint/tendermint => github.com/okex/tendermint v0.32.10-okexchain
)
