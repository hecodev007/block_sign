module rsksync

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cosmos/cosmos-sdk v0.34.4-0.20191010193331-18de630d0ae1
	github.com/dghubble/sling v1.3.0
	github.com/eoscanada/eos-go v0.9.0
	github.com/ethereum/go-ethereum v1.9.15
	github.com/fioprotocol/fio-go v1.0.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/gorilla/websocket v1.4.1
	//github.com/group-coldwallet/common v0.0.0-00010101000000-000000000000
	github.com/jinzhu/gorm v1.9.12
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/tendermint/tendermint v0.32.7
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/sjson v1.1.1
	github.com/walletam/rabbitmq v0.0.0-20220512054530-4de6f6ac8db1
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20200425230154-ff2c4b7c35a0
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0 // indirect
	xorm.io/xorm v0.8.1
)

//replace github.com/group-coldwallet/common => ../common
