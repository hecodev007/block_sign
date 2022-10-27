module dataserver

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d // indirect
	github.com/bwmarrin/snowflake v0.3.0
	github.com/dghubble/sling v1.3.0
	github.com/ethereum/go-ethereum v1.9.15
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/sirupsen/logrus v1.8.1
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/tebeka/strftime v0.1.3 // indirect
	github.com/walletam/rabbitmq v0.0.0-20220512054530-4de6f6ac8db1
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210510120150-4163338589ed
	google.golang.org/appengine v1.6.5
	gopkg.in/yaml.v2 v2.2.4 // indirect
	xorm.io/xorm v0.8.1

)

//replace github.com/group-coldwallet/common => ../common
