package chia

import (
	"bytes"
	"chiaSign/common/log"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func Test_acc(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:8555", "./private_full_node.crt", "./private_full_node.key")
	state, err := rpc.State()
	if err != nil {
		panic(err.Error())
	}
	t.Log(state.BlockchainState.Peak.Height)

}
func Test_wallet_Generatenemonic(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	mo, err := rpc.GenerateMnemonic()
	if err != nil {
		log.Fatal(err.Error())
	}
	monic := strings.Join(mo.Mnemonic, " ")
	t.Log(monic, String(mo))
}
func Test_wallet_getpublickeys(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	pubkeys, err := rpc.GetPublicKeys()

	if err != nil {
		log.Fatal(err.Error())
	}
	t.Log(String(pubkeys))
	//3149734674
}

func Test_wallet_addmon(t *testing.T) {
	decimal.NewFromInt(10).BigInt().Bytes()
	//power prefer just gather flip front round lottery ball seminar weapon present swim label april confirm satisfy middle attend reopen feel match upper march
	monic := "power prefer just gather flip front round lottery ball seminar weapon present swim label april confirm satisfy middle attend reopen feel match upper march"
	//2537826843
	//monic = "wealth cash organ hedgehog swear core above rich clean inspire slender crane phone engage raven predict lady early arrow brush powder glare aware hat"
	//1922074172
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")

	id, err := rpc.AddMonic(monic)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(id)
}
func Test_wallet_get_address(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	t.Log(rpc.Get_next_address(1))
	//xch14xlhx28xad689729jhgd5yxy5z3q7tjy4n2gz0xjnkga6ce350kska7pnm
	//xch1tsmyldg33ml5ls55yfdls8a0u03cu50k0qkqu2wvx3utz6eem6yq0f0jtk
}

func Test_login(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	t.Log(rpc.Login(2537826843))
}
func Test_walletInfo(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	t.Log(rpc.WalletInfo())
}
func Test_walletbalance(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	balance, err := rpc.WalletBalance(1)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(balance))
}

func Test_DelAllKey(t *testing.T) {
	rpc := NewRpcClient("https://chia.net:9256", "./private_wallet.crt", "./private_wallet.key")
	t.Log(rpc.DelAllKey())
}

func String(d interface{}) string {
	data, _ := json.Marshal(d)
	return string(data)
}
func Test_client(t *testing.T) {
	pool := x509.NewCertPool()
	caCertPath := "./private_full_node.crt"

	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt)

	cliCrt, err := tls.LoadX509KeyPair("./private_full_node.crt", "./private_full_node.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			Certificates:       []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}
	params := "{\"\":\"\"}"

	resp, err := client.Post("https://chia.net:8555/get_blockchain_state", "application/json", bytes.NewReader([]byte(params)))
	if err != nil {
		fmt.Println("Get error:", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
