module github.com/group-coldwallet/trxsync

go 1.15

require github.com/BurntSushi/toml v0.3.1

require (
	github.com/Dipper-Labs/Dipper-Protocol v0.0.0-20201103114409-9e306c5ed78f
	github.com/JFJun/trx-sign-go v1.0.0
	github.com/bwmarrin/snowflake v0.3.0
	github.com/fatih/structs v1.1.0
	github.com/fbsobreira/gotron-sdk v0.0.0-20201030191254-389aec83c8f9
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/websocket v1.4.2
	github.com/group-coldwallet/common v0.0.0-20191213075028-77177f63a8fd
	github.com/jinzhu/gorm v1.9.16
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/prometheus/client_golang v1.7.0 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/tendermint/tendermint v0.32.13
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	google.golang.org/genproto v0.0.0-20200831141814-d751682dd103 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	xorm.io/xorm v1.0.5
)

replace github.com/group-coldwallet/common => ../common
