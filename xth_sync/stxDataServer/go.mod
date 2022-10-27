module stxDataServer

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/algorand/go-algorand-sdk v1.5.1
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cosmos/cosmos-sdk v0.39.2
	github.com/gin-gonic/gin v1.5.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.16
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/okex/okexchain v0.16.3
	github.com/okex/okexchain-go-sdk v0.16.0
	github.com/shopspring/decimal v1.2.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/tendermint/tendermint v0.33.9
	github.com/tidwall/gjson v1.6.7
	go.uber.org/zap v1.15.0
	xorm.io/xorm v0.8.1
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/okex/cosmos-sdk v0.39.2-okexchain2
	github.com/tendermint/tendermint => github.com/okex/tendermint v0.33.9-okexchain1
)
