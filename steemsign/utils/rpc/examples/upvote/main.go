package main

import (
	"encoding/json"
	// Stdlib
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"steemsign/utils/rpc/transports/rpcclient"
	"syscall"
	// RPC
	"github.com/robertkrimen/otto"
	"steemsign/utils/rpc"
	"steemsign/utils/rpc/transactions"
	"steemsign/utils/rpc/transports/websocket"
	"steemsign/utils/rpc/types"
	// Vendor
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func run() (err error) {
	// Process flags.
	//host := "http://10.0.230.86:8090"
	host := "https://api.steemit.com"
	flagAddress := flag.String("rpc_endpoint", host, "steemd RPC endpoint address")
	flag.Parse()

	url := *flagAddress
	//brew install automake
	//brew install libtool
	//speck 安装
	//$ ./autogen.sh
	//$ ./configure --enable-module-recovery
	//$ make
	//$ make check
	//$ sudo make install

	// Process args.
	//args := flag.Args()
	//if len(args) != 3 {
	//	return errors.New("3 arguments required")
	//}
	//	author, permlink, voter := args[0], args[1], args[2]

	// Prompt for WIF.
	//wifKey, err := promptWIF(voter)
	//if err != nil {
	//	return err
	//}

	// Start catching signals.
	var interrupted bool
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Drop the error in case it is a request being interrupted.
	defer func() {
		if err == websocket.ErrClosing && interrupted {
			err = nil
		}
	}()

	// Instantiate the WebSocket transport.
	//t, err := websocket.NewTransport([]string{url})
	//if err != nil {
	//	return err
	//}
	t := rpcclient.NewRpcClient(url)

	// Use the transport to get an RPC client.
	// Use the transport to get an RPC client.
	client, err := rpc.NewClient(t)
	if err != nil {
		return err
	}
	defer func() {
		if !interrupted {
			client.Close()
		}
	}()

	// Start processing signals.
	go func() {
		<-signalCh
		fmt.Println()
		log.Println("Signal received, exiting...")
		signal.Stop(signalCh)
		interrupted = true
		client.Close()
	}()

	//client.Database.GetChainPropertiesRaw()
	//	fmt.Printf("%+v\n", resp)
	// Get the props to get the head block number and ID
	// so that we can use that for the transaction.
	props, err := client.Database.GetDynamicGlobalProperties()
	if err != nil {
		return err
	}

	fmt.Println("Id:", props.ID)
	//fmt.Printf("%+v\n", props)

	fmt.Println("begein RefBlockPrefix")
	// Prepare the transaction.
	refBlockPrefix, err := transactions.RefBlockPrefix(props.HeadBlockID)
	if err != nil {
		return err
	}
	fmt.Println("begein NewSignedTransaction")
	//refBlockNum,prefix,fromAccount,toAccount,amountToSend,privateKey
	num := fmt.Sprintf("%v", props.HeadBlockNumber)
	prix := fmt.Sprintf("%v", refBlockPrefix)
	//sign_offline_transfer.js
	//SignTrans(num, prix, "marjay", "tipu", "0.001", "5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")

	jsfile := "/Users/dnsb01357/js/offline_signing/sign_offline_transfer.js"

	fmt.Println(num, prix, jsfile)
	//Command("/bin/bash", "-c", s)
	//ex := time.Now().Add(30 * time.Second)

	//expire := &types.Time{
	//	&ex,
	//}

	//s := fmt.Sprintf("node /Users/dnsb01357/js/offline_signing/sign_offline_transfer.js 46030 2055430191 marjay tipu '0.001' 5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
	s := fmt.Sprintf("node %s %v %v %s %s '%v' %s", jsfile, transactions.RefBlockNum(props.HeadBlockNumber), refBlockPrefix, "marjay", "tipu", "0.002", "5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
	fmt.Println(s)
	//cmd := exec.Command("/bin/bash", "-c", s)
	cmd := exec.Command("/bin/bash", "-c", s)
	//cmd := exec.Command("/bin/bash", "-c", "node", jsfile, num, prix, "marjay", "tipu", "0.001", "5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")

	//cmd := exec.Command("node", jsfile, num, prix, "marjay", "tipu", "0.001", "5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//err = cmd.Run()
	//fmt.Println(err)
	//// 执行命令，并返回结果
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Output->", err)
	}
	fmt.Println("exec output:", string(output))
	/*
		tx := transactions.NewSignedTransaction(&types.Transaction{
			RefBlockNum:    transactions.RefBlockNum(props.HeadBlockNumber),
			RefBlockPrefix: refBlockPrefix,
			Expiration:     expire,
			Extensions:     []string{},
		})

		//tx.PushOperation(&types.VoteOperation{
		//	Voter:    voter,
		//	Author:   author,
		//	Permlink: permlink,
		//	Weight:   10000,
		//})

		tx.PushOperation(&types.TransferOperation{
			From: "marjay",
			To:   "tipu",
			//Amount: "2.124 SBD",
			Amount: "0.001 STEEM",
			Memo:   "",
		})
		//	extensions
		//fmt.Printf("%+v\n", *tx.Transaction)
		// Sign.
		//	privKey, err := wif.Decode(wifKey)
		p := "5J4MA5JJ2M5fW3qovQ1E5pB4UYfZaEVsrDy5ZwtcWP7cY82vAZH"
		//p := "5JLw5dgQAx6rhZEgNN5C2ds1V47RweGshynFSWFbaMohsYsBvE8"
		fmt.Println("wif", len(p))
		privKey, err := wif.Decode(p)
		if err != nil {
			fmt.Println("wif.Decode:", err)
			return err
		}

		activePrivKey, err := wif.Decode("5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
		if err != nil {
			fmt.Println("wif.Decode:", err)
			return err
		}

		memoPrivKey, err := wif.Decode("5KRMYrG29HeVSbZuXuTdrtUtAfMKViwZeuKxqpjFgTFLepjqsqh")
		if err != nil {
			fmt.Println("wif.Decode:", err)
			return err
		}
		fmt.Println("pri len:", len(privKey))

		postPrivKey, err := wif.Decode("5Ke3PUfnVUpyZzC4yQMS7ATELsUDJyfhKoKQyoLbarEc6wg5E9T")
		if err != nil {
			fmt.Println("wif.Decode:", err)
			return err
		}
		fmt.Println(len(memoPrivKey), len(postPrivKey))

		privKeys := [][]byte{activePrivKey}

		if err := tx.Sign(privKeys, transactions.SteemChain); err != nil {
			fmt.Println("sign error:", err)
			return err
		}
		fmt.Println("go->sig", tx.Signatures)

		tx.Signatures = []string{string(output[:len(output)-1])}
	*/
	//pubkey, err := wif.GetPublicKey("5HwTWjvH52LW2pdkAvQo3gYsj1DhFRo81sn88TVFPPbzibKdChS")
	//if err != nil {
	//	fmt.Println("wif.Decode:", err)
	//	return err
	//}
	//
	//b, err := tx.Verify([][]byte{pubkey}, transactions.SteemChain)
	//if err != nil {
	//	fmt.Println("wif.Verify:", err)
	//	return err
	//}
	//
	//fmt.Println("verify:", b)
	// Broadcast.

	trans := &types.Transaction{}
	json.Unmarshal(output, trans)
	resp, err := client.NetworkBroadcast.BroadcastTransactionSynchronous(trans)
	//err = client.NetworkBroadcast.BroadcastTransaction(trans)
	if err != nil {
		fmt.Println("broad cast error:", err)
		return err
	}
	fmt.Printf("%+v\n", *resp)

	// Success!
	return nil
}

//refBlockNum,prefix,fromAccount,toAccount,amountToSend,privateKey
func SignTrans(refBlockNum, prefix, fromAccount, toAccount, amountToSend, privateKey string) {
	jsfile := "/Users/dnsb01357/js/offline_signing/sign_transfer.js"
	bytes, err := ioutil.ReadFile(jsfile)
	if err != nil {
		fmt.Println("readfile error", err)
		return
	}
	vm := otto.New()
	_, err = vm.Run(string(bytes))
	if err != nil {
		fmt.Println("vm run->", err)
		return
	}

	enc, err := vm.Call("signTransaction", nil, refBlockNum, prefix, fromAccount, toAccount, amountToSend, privateKey)
	fmt.Println(enc)
}

func promptWIF(accountName string) (string, error) {
	fmt.Printf("Please insert WIF for account @%v: ", accountName)
	passwd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", errors.Wrap(err, "failed to read WIF from the terminal")
	}
	fmt.Println()
	return string(passwd), nil
}
