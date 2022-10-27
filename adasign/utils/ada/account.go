package ada

import (
	"adasign/common/conf"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/client"

	"github.com/coinbase/rosetta-sdk-go/keys"
	"github.com/coinbase/rosetta-sdk-go/types"

	libada "github.com/Bitrue-exchange/libada-go"

	cardanogo "github.com/onethefour/cardano-go"
)

var wallet *cardanogo.Wallet
var index int

func init() {
	pri := cardanogo.NewEntropy(160)
	//mnemonic := crypto.NewMnemonic(pri)
	////pri := make([]byte,entropySizeInBits)
	////n,err := rand.Read(pri)
	////if err != nil {
	////	t.Fatal(err)
	////}
	////if n != entropySizeInBits{
	////	t.Fatal()
	////}
	wallet = cardanogo.NewWallet("goapi", "", pri)
	wallet.SetNetwork(cardanogo.Mainnet)
	index = 1
}

func GenAccount() (addr string, pri string, err error) {
	kp, err := keys.GenerateKeypair(types.Edwards25519)
	if err != nil {
		return "", "", err
	}
	address := libada.NewKeyedEnterpriseAddress(kp.PublicKey.Bytes, libada.Mainnet) //这个是主网的
	//address := libada.NewKeyedEnterpriseAddress(kp.PublicKey.Bytes, libada.Testnet) //这个是测试网的
	return address.String(), hex.EncodeToString(kp.PrivateKey), nil
}

func ToKeyPire(seed string) (*keys.KeyPair, error) {
	seedbytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, err
	}
	privateKey := ed25519.NewKeyFromSeed(seedbytes)
	publicKey := make([]byte, ed25519.PublicKeySize)
	copy(publicKey, privateKey[32:])
	pubKey := &types.PublicKey{
		Bytes:     publicKey,
		CurveType: types.Edwards25519,
	}

	kp2 := &keys.KeyPair{
		PublicKey:  pubKey,
		PrivateKey: privateKey.Seed(),
	}
	return kp2, nil
	//addr = libada.NewKeyedEnterpriseAddress(kp2.PublicKey.Bytes, libada.Mainnet)
}
func NewRpcCli(url string) *client.APIClient {
	//ctx := context.Background()
	clientCfg := client.NewConfiguration(
		conf.GetConfig().Node.Url,
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	return cli
}
