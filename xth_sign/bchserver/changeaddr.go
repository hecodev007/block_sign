package main

import (
	"fmt"
	"github.com/gcash/bchd/chaincfg"
	"github.com/gcash/bchutil"
)

func main() {
	//q转1
	addr, _ := bchutil.DecodeAddress("qpfa8wtdan5k8230eaj7kxhvkzajkhvyy522l3h0hc", &chaincfg.MainNetParams)
	addr1, _ := bchutil.NewLegacyAddressPubKeyHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
	//18eEincrd9v5mG3NvxLT1MDpgG94SYXYTT
	fmt.Println(addr1.String())

	//1转q
	addr, _ = bchutil.DecodeAddress("18eEincrd9v5mG3NvxLT1MDpgG94SYXYTT", &chaincfg.MainNetParams)
	addr, _ = bchutil.NewAddressPubKeyHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
	//qpfa8wtdan5k8230eaj7kxhvkzajkhvyy522l3h0hc
	fmt.Println(addr.String())

	//3转p
	addr, _ = bchutil.DecodeAddress("3NFvYKuZrxTDJxgqqJSfouNHjT1dAG1Fta", &chaincfg.MainNetParams)
	addr, _ = bchutil.NewAddressScriptHashFromHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
	//prseh0a4aejjcewhc665wjqhppgwrz2lw5txgn666a
	fmt.Println(addr.String())

	//p转3
	addr, _ = bchutil.DecodeAddress("prseh0a4aejjcewhc665wjqhppgwrz2lw5txgn666a", &chaincfg.MainNetParams)
	addr, _ = bchutil.NewLegacyAddressScriptHashFromHash(addr.ScriptAddress(), &chaincfg.MainNetParams)
	//3NFvYKuZrxTDJxgqqJSfouNHjT1dAG1Fta
	fmt.Println(addr, addr.String(), addr.EncodeAddress())
}
