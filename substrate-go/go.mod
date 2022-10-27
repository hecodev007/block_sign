module github.com/rjman-ljm/substrate-go

go 1.15

replace github.com/centrifuge/go-substrate-rpc-client/v3 v3.0.2 => github.com/chainx-org/go-substrate-rpc-client/v3 v3.1.1

require (
	github.com/centrifuge/go-substrate-rpc-client/v3 v3.0.2
	github.com/ethereum/go-ethereum v1.10.4
	github.com/hacpy/chainbridge-substrate-events v1.0.0
	github.com/huandu/xstrings v1.3.2
	github.com/rjman-ljm/go-substrate-crypto v1.0.0
	github.com/shopspring/decimal v1.2.0
	github.com/stafiprotocol/go-substrate-rpc-client v1.0.5
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
)
