package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"steemsync/utils/rpc"
	"steemsync/utils/rpc/transports/rpcclient"
	"steemsync/utils/rpc/transports/websocket"
	"steemsync/utils/rpc/types"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func run() (err error) {
	// Process flags.
	//https://api.steemit.com
	//http://10.0.230.86:8090
	//

	host := "https://api.steemit.com"
	//host := "http://10.0.230.86:8090"
	flagAddress := flag.String("rpc_endpoint", host, "steemd RPC endpoint address")
	flagReconnect := flag.Bool("reconnect", false, "enable auto-reconnect mode")
	flag.Parse()

	var (
		url       = *flagAddress
		reconnect = *flagReconnect
	)

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

	// Start the connection monitor.
	monitorChan := make(chan interface{}, 1)
	if reconnect {
		go func() {
			for {
				event, ok := <-monitorChan
				if ok {
					log.Println(event)
				}
			}
		}()
	}

	// Instantiate the WebSocket transport.
	log.Printf("---> Dial(\"%v\")\n", url)
	//t, err := websocket.NewTransport([]string{url},
	//	websocket.SetAutoReconnectEnabled(reconnect),
	//	websocket.SetAutoReconnectMaxDelay(30*time.Second),
	//	websocket.SetMonitor(monitorChan))
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}

	t := rpcclient.NewRpcClient(url)

	// Use the transport to get an RPC client.
	client, err := rpc.NewClient(t)
	if err != nil {
		fmt.Println("NewClient:", err)
		return err
	}

	//info, err := client.Database.GetTransactionInfo("47ebbcc58183804521be6eb320ecfd6be31f2cba")
	////fmt.Println(info)
	//fmt.Printf("%+v\n", *info)
	//
	//status, err := client.Database.GetTransactionStatus("47ebbcc58183804521be6eb320ecfd6be31f2cba")
	//fmt.Printf("%+v\n", *status)
	fmt.Println("get client end")
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

	// Get config.
	log.Println("---> GetConfig()")
	config, err := client.Database.GetConfig()
	if err != nil {
		fmt.Println("Database.GetConfig()", err)
		return err
	}

	fmt.Printf("config：%+v", config)

	object, err := client.Database.GetOpsInBlock(65211016, false)
	if err != nil {
		fmt.Println(err)

	}
	for _, o := range object {
		switch op := o.Operation.Data().(type) {
		case *types.TransferOperation:
			if op.From == "steemgoapi" || op.To == "steemgoapi" {
				fmt.Printf("Transfer from %v,to %v,memo %v,amount %v\n", op.From, op.To, op.Memo, op.Amount)
			}

		}
	}
	//operation := tx.Operation.Data().(*types.TransferOperation)
	//amounts := strings.Split(operation.Amount, "")
	//amount, _ := decimal.NewFromString(amounts[0])
	// Use the last irreversible block number as the initial last block number.
	props, err := client.Database.GetDynamicGlobalProperties()
	if err != nil {
		return err
	}
	fmt.Println("total active steem->", props.TotalActivityFundSteem)

	lastBlock := props.LastIrreversibleBlockNum

	// Keep processing incoming blocks forever.
	log.Printf("---> Entering the block processing loop (last block = %v)\n", lastBlock)
	for {
		// Get current properties.
		props, err := client.Database.GetDynamicGlobalProperties()
		if err != nil {
			return err
		}

		// Process new blocks.
		for props.LastIrreversibleBlockNum-lastBlock > 0 {
			object, err := client.Database.GetOpsInBlock(lastBlock, false)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, o := range object {
				switch op := o.Operation.Data().(type) {
				case *types.TransferOperation:
					fmt.Printf("Transfer from %v,to %v,memo %v,amount %v\n", op.From, op.To, op.Memo, op.Amount)
				}
			}
			/*	block, err := client.Database.GetBlock(lastBlock)
				if err != nil {
					return err
				}

				// Process the transactions.
				for _, tx := range block.Transactions {
					for _, operation := range tx.Operations {
						switch op := operation.Data().(type) {
						case *types.VoteOperation:
							//fmt.Printf("@%v voted for @%v/%v\n", op.Voter, op.Author, op.Permlink)

						case *types.TransferOperation:
							fmt.Printf("Transfer from %v,to %v,memo %v,amount %v\n", op.From, op.To, op.Memo, op.Amount)
							fmt.Println(op.Data())

							// You can add more cases here, it depends on
							// what operations you actually need to process.
						}
					}
				}
			*/
			lastBlock++
		}

		// Sleep for STEEMIT_BLOCK_INTERVAL seconds before the next iteration.
		time.Sleep(time.Duration(config.SteemitBlockInterval) * time.Second)
	}
}
