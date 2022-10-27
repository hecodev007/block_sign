module lunasync

go 1.13

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/cosmos/ibc-go v1.2.5 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.12
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/o3labs/neo-utils v0.0.0-20190806035218-cbe201aea47a
	github.com/onethefour/common v0.0.4
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/streadway/amqp v1.0.0
	github.com/tendermint/tendermint v0.34.15
	github.com/terra-money/core v0.5.16
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20211005001312-d4b1ae081e3b
	google.golang.org/appengine v1.6.7
	xorm.io/xorm v0.8.1

)

replace (
	// Use patched version based on v0.44.5 - note: not state compatiable
	github.com/cosmos/cosmos-sdk => github.com/kava-labs/cosmos-sdk v0.44.5-kava.1
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/tecbot/gorocksdb => github.com/cosmos/gorocksdb v1.2.0
	github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.15
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
