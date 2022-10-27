module bncsync

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/JFJun/bifrost-go v1.0.4
	github.com/JFJun/go-substrate-crypto v1.0.1
	github.com/JFJun/go-substrate-rpc-client/v3 v3.0.3
	github.com/JFJun/substrate-go v1.2.3
	github.com/SubGame-Network/go-substrate-rpc-client v1.1.0
	github.com/bwmarrin/snowflake v0.3.0
	github.com/centrifuge/go-substrate-rpc-client v2.0.0+incompatible
	github.com/centrifuge/go-substrate-rpc-client/v3 v3.0.2
	github.com/gin-gonic/gin v1.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.12
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/o3labs/neo-utils v0.0.0-20190806035218-cbe201aea47a
	github.com/onethefour/bifrost-go v1.0.8
	github.com/onethefour/common v0.0.4
	github.com/shopspring/decimal v1.2.0
	github.com/stafiprotocol/go-substrate-rpc-client v1.1.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	go.uber.org/zap v1.19.0
	xorm.io/xorm v0.8.1
)

replace github.com/centrifuge/go-substrate-rpc-client => ../../../go/src/github.com/centrifuge/go-substrate-rpc-client

replace github.com/JFJun/bifrost-go => ../../../go/src/github.com/JFJun/bifrost-go

//replace github.com/JFJun/go-substrate-rpc-client => ../../../go/src/github.com/JFJun/go-substrate-rpc-client
