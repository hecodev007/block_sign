package waves

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/onethefour/common/xutils"
	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
	"github.com/wavesplatform/gowaves/pkg/settings"
	"golang.org/x/net/context"
)

func Test_addr(t *testing.T) {
	var scheme = settings.MainNetSettings.AddressSchemeCharacter
	var seed [32]byte
	_, err := io.ReadFull(rand.Reader, seed[:])
	if err != nil {
		t.Fatal(err.Error())
	}
	privkey, pubkey, err := crypto.GenerateKeyPair(seed[:])
	_, _ = privkey, pubkey
	t.Log(pubkey.String())
	addr, err := proto.NewAddressFromPublicKey(scheme, pubkey)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(addr.String())
}

func Test_pub(t *testing.T) {
	seed, _ := hex.DecodeString("3b72702ae3f4798ec160aa6ea479b1f636cc805ba0344b38737d444ba0778326")
	privkey, _, _ := crypto.GenerateKeyPair(seed[:])
	t.Log(privkey.String())
	////var scheme byte = 'T'
	//pub := "GBKSAe9DBHdsqrTFZkDrvqDR5cAzHmRGqGxYL59q8VxE"
	//pubkey, err := crypto.NewPublicKeyFromBase58(pub)
	//if err != nil {
	//	t.Fatal(err.Error())
	//}
	//for i := uint8(0); i < math.MaxUint8; i++ {
	//	addr, _ := proto.NewAddressFromPublicKey(i, pubkey)
	//	t.Log(i, addr.String())
	//}

}

func Test_client(t *testing.T) {
	cli, err := client.NewClient()
	if err != nil {
		t.Fatal(err.Error())
	}
	address, err := proto.NewAddressFromString("3PMEHLx1j6zerarZTYfsGqDeeZqQoMpxq5S")
	if err != nil {
		t.Fatal(err.Error())
	}
	balance, resp, err := cli.Addresses.Balance(context.Background(), address)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(balance))
	t.Log(xutils.String(resp))
	asertBalance, _, err := cli.Assets.BalanceByAddress(context.Background(), address)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(asertBalance))
}

func Test_sign(t *testing.T) {
	cli, err := client.NewClient()
	if err != nil {
		t.Fatal(err.Error())
	}
	//addr :="3PMEHLx1j6zerarZTYfsGqDeeZqQoMpxq5S"
	address, err := proto.NewAddressFromString("3PMEHLx1j6zerarZTYfsGqDeeZqQoMpxq5S")
	if err != nil {
		t.Fatal(err.Error())
	}
	waves := proto.NewOptionalAssetWaves()
	var pk crypto.PublicKey
	tx := proto.NewUnsignedTransferWithSig(pk, waves, waves, 100, 1, 100, proto.NewRecipientFromAddress(address), []byte("attachment"))
	reponse, err := cli.Transactions.Broadcast(context.Background(), tx)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(reponse))
	t.Log(tx.ID.String())
}
