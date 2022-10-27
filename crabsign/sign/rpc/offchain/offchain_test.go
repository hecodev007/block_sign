package offchain

import (
	"os"
	"testing"

	"wallet-sign/sign/client"
	"wallet-sign/sign/config"
)

var offchain *Offchain

func TestMain(m *testing.M) {
	cl, err := client.Connect(config.Default().RPCURL)
	if err != nil {
		panic(err)
	}
	offchain = NewOffchain(cl)
	os.Exit(m.Run())
}
