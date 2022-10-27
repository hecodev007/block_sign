module solsign

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/adiabat/bech32 v0.0.0-20170505011816-6289d404861d
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/dfuse-io/solana-go v0.2.0
	github.com/ethereum/go-ethereum v1.9.24
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/onethefour/common v0.0.3
	github.com/pkg/errors v0.9.1 // indirect
	github.com/portto/solana-go-sdk v1.1.0
	github.com/shopspring/decimal v1.2.0
	github.com/streamingfast/solana-go v0.0.0-00010101000000-000000000000
	github.com/tebeka/strftime v0.1.5 // indirect
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/exp v0.0.0-20201008143054-e3b2a7f2fdc7
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	google.golang.org/appengine v1.6.5
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/streamingfast/solana-go => ../../github/mainchain/solana-go
