package fio

import (
	"testing"

	"github.com/fioprotocol/fio-go"
)

func Test_cli(t *testing.T) {
	wif := "5JP1fUXwPxuKuNryh5BEsFhZqnh59yVtpHqHxMMTmtjcni48bqC"
	url := "https://fio.greymass.com"
	key, err := fio.NewAccountFromWif(wif)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(key)
	account, api, _, err := fio.NewWifConnect(wif, url)
	t.Log(account)
	api.GetFioAddresses()
}

//nodeUrl = "https://fio.greymass.com"
