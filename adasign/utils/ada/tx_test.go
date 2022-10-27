package ada

import (
	"encoding/hex"
	"testing"

	cardanogo "github.com/onethefour/cardano-go"
)

func Test_tx(t *testing.T) {
	pri, _ := hex.DecodeString("9ec094999e79ff457c19614e6d31323af63b9574")
	wa := cardanogo.NewWallet("", "", pri)
	wa.SetNetwork(cardanogo.Mainnet)
	t.Log(wa.Addresses())

}

//addr1vxwmn225swwsq8uqtg8f7t7w0sn0ggx7pfaeacwxseyx2yqmf4hmq 9ec094999e79ff457c19614e6d31323af63b9574
//addr1v9se77w3gvr90qspc6f9t4t3wwgj8asn84cuqgu8uqgyk9cq34l67 faa2390922ea2532032ac1ec719839d8ec20aefd
//addr1v832ehavrtrr925kzuzlvkwmnyrk8ascz4qe22zef8lgskq4c93a9 c6096d7153a676f522fbe2fd0f56bb6a0c06d539
