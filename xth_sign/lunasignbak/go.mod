module terrasign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	//github.com/cosmos/cosmos-sdk v0.42.5
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/gin-gonic/gin v1.4.0
	github.com/kava-labs/kava v0.16.1
	//github.com/kava-labs/rosetta-kava v0.0.0-20210910213328-57ccd4fe3063
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/onethefour/common v0.0.4
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/tendermint/tendermint v0.34.14 // indirect
	github.com/terra-project/terra.go v1.0.0
	github.com/ugorji/go v1.1.7 // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/exp v0.0.0-20200513190911-00229845015e
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	google.golang.org/appengine v1.6.6

)

//github.com/cosmos/cosmos-sdk v0.39.2
//github.com/tendermint/tendermint v0.33.9
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

//replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.39.2

//replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.33.9

//replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.32.7
//replace github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.34.4-0.20191010193331-18de630d0ae1
