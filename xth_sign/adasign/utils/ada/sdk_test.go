package ada

import (
	"context"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/onethefour/common/xutils"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/Bitrue-exchange/libada-go"

	"github.com/coinbase/rosetta-sdk-go/client"
)

func Test_sdk(t *testing.T) {
	ctx := context.Background()
	clientCfg := client.NewConfiguration(
		"http://54.250.240.45:8080",
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)

	_ = cli
	seed := "83d02ce18dc4764144edc1d3acf936ba5d759c3c1b8177317f0608ad4e7b77e7"
	kp, err := ToKeyPire(seed)
	if err != nil {
		t.Fatal(err.Error())
	}

	address := libada.NewKeyedEnterpriseAddress(kp.PublicKey.Bytes, libada.Mainnet)
	t.Log(address)
	preprocessRequest := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: NetworkIdentifier,
		Operations:        testOPs(),
	}
	t.Log(xutils.String(testOPs()))
	preresp, cerr, err := cli.ConstructionAPI.ConstructionPreprocess(ctx, preprocessRequest)
	if err != nil {
		t.Fatal(err.Error())
	}
	if cerr != nil {
		t.Fatal(cerr.Message)
	}

	//cli.NetworkAPI.NetworkOptions()
	req := &types.ConstructionMetadataRequest{
		NetworkIdentifier: &types.NetworkIdentifier{
			Blockchain: "cardano",
			Network:    "mainnet",
		},
		Options:    preresp.Options,
		PublicKeys: []*types.PublicKey{kp.PublicKey},
	}
	MetaDataresp, terr, err := cli.ConstructionAPI.ConstructionMetadata(context.Background(), req)
	if err != nil {
		t.Fatal(err.Error())
	}
	if terr != nil {
		t.Fatal(terr.Message)
	}
	t.Log(xutils.String(MetaDataresp))
	payloadRrqust := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: NetworkIdentifier,
		Operations:        testOPs(),
		Metadata:          MetaDataresp.Metadata,
	}
	payloadresp, cerr, err := cli.ConstructionAPI.ConstructionPayloads(ctx, payloadRrqust)
	if err != nil {
		t.Fatal(err.Error())
	}
	if terr != nil {
		t.Fatal(terr.Message)
	}
	t.Log(xutils.String(payloadresp))
	t.Log(seed)
	t.Log(hex.EncodeToString(kp.PrivateKey))
	signer, _ := kp.Signer()
	sigerbytes, err := signer.Sign(payloadresp.Payloads[0], "ed25519")
	combileRequest := &types.ConstructionCombineRequest{
		NetworkIdentifier:   NetworkIdentifier,
		UnsignedTransaction: payloadresp.UnsignedTransaction,
		Signatures: []*types.Signature{&types.Signature{
			SigningPayload: payloadresp.Payloads[0],
			PublicKey:      kp.PublicKey,
			SignatureType:  "ed25519",
			Bytes:          sigerbytes.Bytes,
		}},
	}

	combileResp, cerr, err := cli.ConstructionAPI.ConstructionCombine(ctx, combileRequest)
	if err != nil {
		t.Fatal(err.Error())
	}
	if terr != nil {
		t.Fatal(terr.Message)
	}
	t.Log(xutils.String(combileResp))
	//cli.ConstructionAPI.ConstructionCombine()
	//cli.ConstructionAPI.ConstructionDerive(context.Background())
	//cli.ConstructionAPI.ConstructionPayloads()
}

func Test_input(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	clientCfg := client.NewConfiguration(
		"http://54.250.240.45:8080",
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	totalInput := make(map[string]uint64)
	op, err := NewInputOperation(cli, "4125f549c83f2606323e555883fbfefccc4fc68e142bd89df385071b6200cbb9", 0, totalInput)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(op))

	toaddr := "addr1q8e95sy56afr8kj7t6czy6qx6ztghe038p40spwa0xft54wc0h7ch26lev50xa9508swl04n7epcvw7p82cvfgv9xmkqhuzq6c"
	adamount := uint64(1000000)
	outop, err := NewOutputOperation(toaddr, adamount, "", 0)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(xutils.String(outop))
}
func testOPs() (ret []*types.Operation) {
	ctx := context.Background()
	_ = ctx
	clientCfg := client.NewConfiguration(
		"http://54.250.240.45:8080",
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	totalInput := make(map[string]uint64)
	inop, err := NewInputOperation(cli, "cfc9c02db1421f077fd03e8d46895e11710f528243c7a77e570dc15533aa3e96", 0, totalInput)
	if err != nil {
		panic(err.Error())
	}

	toaddr := "addr1q8e95sy56afr8kj7t6czy6qx6ztghe038p40spwa0xft54wc0h7ch26lev50xa9508swl04n7epcvw7p82cvfgv9xmkqhuzq6c"
	adamount := uint64(1330000)
	outop, err := NewOutputOperation(toaddr, adamount, "", 0)
	if err != nil {
		panic(err.Error())
	}
	ret = []*types.Operation{inop, outop}
	return ret
}

func Test_acc2(t *testing.T) {
	seed := "83d02ce18dc4764144edc1d3acf936ba5d759c3c1b8177317f0608ad4e7b77e7"
	kp, err := ToKeyPire(seed)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(hex.EncodeToString(kp.PrivateKey), hex.EncodeToString(kp.PublicKey.Bytes), kp.PublicKey.CurveType)
	addr := libada.NewKeyedEnterpriseAddress(kp.PublicKey.Bytes, libada.Mainnet)
	t.Log(addr.String())
}

func Test_block(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	clientCfg := client.NewConfiguration(
		"http://54.250.240.45:8080",
		"rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)
	cli := client.NewAPIClient(clientCfg)
	index := int64(6576039)
	req := &types.BlockRequest{
		NetworkIdentifier: NetworkIdentifier,
		BlockIdentifier:   &types.PartialBlockIdentifier{Index: &index},
	}
	block, cerr, err := cli.BlockAPI.Block(ctx, req)
	if err != nil {
		t.Fatal(err.Error())
	} else if cerr != nil {
		t.Fatal(cerr.Message)
	}
	t.Log(block.OtherTransactions)
}
