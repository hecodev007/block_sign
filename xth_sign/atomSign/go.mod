module atomSign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/cosmos/cosmos-sdk v0.41.3
	github.com/gin-gonic/gin v1.4.0
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/onethefour/common v0.0.4
	github.com/shopspring/decimal v1.2.0
	github.com/tendermint/tendermint v0.34.7
	github.com/ugorji/go v1.1.7 // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	google.golang.org/appengine v1.6.6
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
