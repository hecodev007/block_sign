package avax

import (
	"encoding/json"
	"fmt"
	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/utils/codec"
	"github.com/ava-labs/gecko/vms/avm"
	"github.com/ava-labs/gecko/vms/secp256k1fx"
	"testing"

	"github.com/ava-labs/gecko/utils/formatting"
	"github.com/ava-labs/gecko/vms/spdagvm"
)

func Test_gnAccount(t *testing.T) {
	u := "11CcxUktd1KxgTstPCCCm2hz6gKhuRhJDCFE6vkQhmFt6FBh93sLtjWnxRZ3eXEW9Cmz5b7WP26kspTPCVRNpRGGYfGEbadGsWqbahC2kgnGpx2Qa6jLLfKTbhqLLHB1pn4thfNhSjw2zjRmFurydMDYcTKNAktuBNxhE1"
	utxo, err := ParseUtxo(u)
	if err != nil {
		panic(err.Error())
	}
	utxojson, _ := json.Marshal(utxo)
	fmt.Println(string(utxojson))
}

// 	pri, addr, err := GenAccount()
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	}
// 	fmt.Println(pri, addr)

// 	shotid, prefix, err := AddressToShot(addr)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	}
// 	fmt.Println(prefix, "shotid:", shotid.Hex())

// 	w, err := NewWallet(pri)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	stid := w.GetAddress()

// 	fmt.Println(prefix, "shotid:", stid.Hex())
// 	// [
// 	// 	"11PQ1sNw9tcXjVki7261souJnr1TPFrdVCu5JGZC7Shedq3a7xvnTXkBQ162qMYxoerMdwzCM2iM1wEQPwTxZbtkPASf2tWvddnsxPEYndVSxLv8PDFMwBGp6UoL35gd9MQW3UitpfmFsLnAUCSAZHWCgqft2iHKnKRQRz",
// 	// 	"11RCDVNLzFT8KmriEJN7W1in6vB2cPteTZHnwaQF6kt8B2UANfUkcroi8b8ZSEXJE74LzX1mmBvtU34K6VZPNAVxzF6KfEA8RbYT7xhraioTsHqxVr2DJhZHpR3wGWdjUnRrqSSeeKGE76HTiQQ8WXoABesvs8GkhVpXMK",
// 	// 	"11GxS4Kj2od4bocNWMQiQhcBEHsC3ZgBP6edTgYbGY7iiXgRVjPKQGkhX5zj4NC62ZdYR3sZAgp6nUc75RJKwcvBKm4MGjHvje7GvegYFCt4RmwRbFDDvbeMYusEnfVwvpYwQycXQdPFMe12z4SP4jXjnueernYbRtC4qL",
// 	// 	"11S1AL9rxocRf2NVzQkZ6bfaWxgCYch7Bp2mgzBT6f5ru3XEMiVZM6F8DufeaVvJZnvnHWtZqocoSRZPHT5GM6qqCmdbXuuqb44oqdSMRvLphzhircmMnUbNz4TjBxcChtks3ZiVFhdkCb7kBNLbBEmtuHcDxM7MkgPjHw",
// 	// 	"11Cn3i2T9SMArCmamYUBt5xhNEsrdRCYKQsANw3EqBkeThbQgAKxVJomfc2DE4ViYcPtz4tcEfja38nY7kQV7gGb3Fq5gxvbLdb4yZatwCZE7u4mrEXT3bNZy46ByU8A3JnT91uJmfrhHPV1M3NUHYbt6Q3mJ3bFM1KQjE"
// 	// ]
// 	assetID := ids.Empty.Prefix(0)
// 	utxo := &avax.UTXO{
// 		UTXOID: avax.UTXOID{TxID: ids.Empty.Prefix(1)},
// 		Asset:  avax.Asset{ID: assetID},
// 		Out: &secp256k1fx.TransferOutput{
// 			Amt: 1000,
// 			OutputOwners: secp256k1fx.OutputOwners{
// 				Threshold: 1,
// 				Addrs:     []ids.ShortID{*shotid},
// 			},
// 		},
// 	}

// 	fb := formatting.CB58{}
// 	acodec := codec.NewDefault()
// 	{
// 		acodec.RegisterType(&secp256k1fx.TransferOutput{})
// 		acodec.RegisterType(&avm.BaseTx{})
// 		acodec.RegisterType(&avm.CreateAssetTx{})
// 		acodec.RegisterType(&avm.OperationTx{})
// 		acodec.RegisterType(&avm.ImportTx{})
// 		acodec.RegisterType(&avm.ExportTx{})
// 		acodec.RegisterType(&secp256k1fx.TransferInput{})
// 		acodec.RegisterType(&secp256k1fx.MintOutput{})
// 		acodec.RegisterType(&secp256k1fx.TransferOutput{})
// 		acodec.RegisterType(&secp256k1fx.MintOperation{})
// 		acodec.RegisterType(&secp256k1fx.Credential{})
//acodec.RegisterType(&ProposalBlock{})
//acodec.RegisterType(&Abort{})
//acodec.RegisterType(&Commit{})
//acodec.RegisterType(&StandardBlock{})
//acodec.RegisterType(&AtomicBlock{})
//
//// The Fx is registered here because this is the same place it is
//// registered in the AVM. This ensures that the typeIDs match up for
//// utxos in shared memory.
//acodec.RegisterType(&secp256k1fx.TransferInput{})
//acodec.RegisterType(&secp256k1fx.MintOutput{})
//acodec.RegisterType(&secp256k1fx.TransferOutput{})
//acodec.RegisterType(&secp256k1fx.MintOperation{})
//acodec.RegisterType(&secp256k1fx.Credential{})
//acodec.RegisterType(&secp256k1fx.Input{})
//acodec.RegisterType(&secp256k1fx.OutputOwners{})
//
//acodec.RegisterType(&UnsignedAddDefaultSubnetValidatorTx{})
//acodec.RegisterType(&UnsignedAddNonDefaultSubnetValidatorTx{})
//acodec.RegisterType(&UnsignedAddDefaultSubnetDelegatorTx{})
//
//acodec.RegisterType(&UnsignedCreateChainTx{})
//acodec.RegisterType(&UnsignedCreateSubnetTx{})
//
//acodec.RegisterType(&UnsignedImportTx{})
//acodec.RegisterType(&UnsignedExportTx{})
//
//acodec.RegisterType(&UnsignedAdvanceTimeTx{})
//acodec.RegisterType(&UnsignedRewardValidatorTx{})
//
//acodec.RegisterType(&StakeableLockIn{})
//acodec.RegisterType(&StakeableLockOut{})
// 	}

// 	//spdcodec :=spdagvm.Codec{}
// 	bts, err := acodec.Marshal(utxo)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	fb.FromString("115ZLnNqzCsyugMY5kbLnsyP2y4se4GJBbKHjyQnbPfRBitqLaxMizsaXbDMU61fHV2MDd7fGsDnkMzsTewULi94mcjk1bfvP7aHYUG2i3XELpV9guqsCtv7m3m3Kg4Ya1m6tAWqT7PhvAaW4D3fk8W1KnXu5JTWvYBqD2")
// 	fmt.Printf("000%x\n", bts)
// 	utxo2 := &avax.UTXO{}
// 	err = acodec.Unmarshal(fb.Bytes, utxo2)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	u2json, err := acodec.Marshal(utxo2)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	fb.Bytes = u2json
// 	fmt.Println(string(u2json))
// 	fmt.Println(fb.String())
// 	//w.AddUtxo(utxo)
// }

func Test_genAccount(t *testing.T) {

	fb := formatting.CB58{}

	fb.FromString("111111111cKwsGT3RZZ7Jms9xUNH6nPS2TUYNeb2ooZxGuJMuC6vU3YMmBJW3eqjTfhTXgmoFHtxmsSQ1UZXWc8FumsRUnNUvxwLN7tF5D6Z1csaYRQBbLVUxyyUA8ZYqBgvLVNCfuKLFcs6wrbZMurVcyDBZi6XHth8PuvdVt6TaQPAoZUmrqnhjmweawzRJtSgrQnYPJFsv1Vf8AbAeav3LK6JZpVdGbXFwoRgFqJYWTNvsjabhFjKCLktFNpcr5WQkFKV6Ti1QMGPRAwC13a86d2ge4ZxpXQNtoCw9cU2SZPpWC6qGRTAqpd8AaTuyisDh89NX2Jdqoq5Znw8wyCsGVUvVeWKtY1p21BqEcmf9BeHi2RbVorQRatqNjzfMeRNbNSdfeuwFzaBMGQD5gVbtW3Lf3inhJSCKZERvdj7JvCdcaMCkesFMgnSJCM3Dritae6MHqjhf1D5wCRf1HvQUNayJxzgBYEDepbcrdnjQJ9zmvy76CAtTVKqqo2ryqC3LSZ6ppb6o9sjEos9npFnvvpqXMKb2GFW6mQNU2LkNcyho2PHTGQExoGiZ7mwxSrYpt4P4PQFagwn73oXnj6rBrEvgrXSD6rAbdj6cKGSUopp2JiU9vEPXTcnCpUgmZNUemj8jCKpMRciRB6AG84CC5tijAykgeafhCjrxE4TKAoVJUxKxxrKDkQqxTVMa4cdzzVC9sExEc6aqZ62kCkM429fYfLGCJRn2CVza9hjtna1tpvqbnEpgTqwYMZjYUBdRgzqZ9ozRXS1Rn3avfkmLYqvqNTK8GHKZAVWgDwLmMSUKWVjCza9pH86MGKWTrR9ouuuLL9Du9Vy")
	//fmt.Printf("%x\n", fb.Bytes)
	//utxo2 := &avax.UTXO{}
	//acodec := platformvm.Codec
	c := codec.NewDefault()
	c.RegisterType(&avm.BaseTx{})
	c.RegisterType(&avm.CreateAssetTx{})
	c.RegisterType(&avm.OperationTx{})
	c.RegisterType(&avm.ImportTx{})
	c.RegisterType(&avm.ExportTx{})
	c.RegisterType(&secp256k1fx.TransferInput{})
	c.RegisterType(&secp256k1fx.MintOutput{})
	c.RegisterType(&secp256k1fx.TransferOutput{})
	c.RegisterType(&secp256k1fx.MintOperation{})
	c.RegisterType(&secp256k1fx.Credential{})

	txBytes := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa8, 0x66,
		0x05, 0x04, 0x03, 0x02, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x70, 0xae, 0x33, 0xb5,
		0x60, 0x9c, 0xd8, 0x9a, 0x72, 0x92, 0x4f, 0xa2,
		0x88, 0x3f, 0x9b, 0xf1, 0xc6, 0xd8, 0x9f, 0x07,
		0x09, 0x9b, 0x2a, 0xd7, 0x1b, 0xe1, 0x7c, 0x5d,
		0x44, 0x93, 0x23, 0xdb, 0x00, 0x00, 0x00, 0x05,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc3, 0x50,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x70, 0xae, 0x33, 0xb5, 0x60, 0x9c, 0xd8, 0x9a,
		0x72, 0x92, 0x4f, 0xa2, 0x88, 0x3f, 0x9b, 0xf1,
		0xc6, 0xd8, 0x9f, 0x07, 0x09, 0x9b, 0x2a, 0xd7,
		0x1b, 0xe1, 0x7c, 0x5d, 0x44, 0x93, 0x23, 0xdb,
		0x00, 0x00, 0x00, 0x01, 0x70, 0xae, 0x33, 0xb5,
		0x60, 0x9c, 0xd8, 0x9a, 0x72, 0x92, 0x4f, 0xa2,
		0x88, 0x3f, 0x9b, 0xf1, 0xc6, 0xd8, 0x9f, 0x07,
		0x09, 0x9b, 0x2a, 0xd7, 0x1b, 0xe1, 0x7c, 0x5d,
		0x44, 0x93, 0x23, 0xdb, 0x00, 0x00, 0x00, 0x05,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc3, 0x50,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x09,
		0x00, 0x00, 0x00, 0x01, 0x50, 0x6b, 0xd9, 0x2d,
		0xe5, 0xeb, 0xc2, 0xbf, 0x8f, 0xaa, 0xf1, 0x7d,
		0xbb, 0xae, 0xb3, 0xf3, 0x13, 0x9e, 0xae, 0xb4,
		0xad, 0x32, 0x95, 0x6e, 0x92, 0x74, 0xf9, 0x53,
		0x0e, 0xcc, 0x03, 0xd8, 0x02, 0xab, 0x1c, 0x16,
		0x52, 0xd0, 0xe3, 0xfc, 0xe5, 0x93, 0xa9, 0x8e,
		0x96, 0x1e, 0x83, 0xf0, 0x12, 0x27, 0x66, 0x9f,
		0x03, 0x56, 0x9f, 0x17, 0x1b, 0xd1, 0x22, 0x90,
		0xfd, 0x64, 0xf5, 0x73, 0x01,
	}
	_ = txBytes
	tx := new(avm.Tx)
	if err := c.Unmarshal(fb.Bytes, &tx); err != nil {
		panic(err.Error())
	}
	//fmt.Println(tx.InputUTXOs()[0].InputID().Hex(), tx.UTXOs()[0].ID.Hex())

}
func Test_ParseTx(t *testing.T) {
	rawTx := "1111111112zs9UkarRq285UVQ1boTn5dMj9VvvaMV9TgjTSGrSq4fJXqJBf7BB9mcFtG98XYDn3nCTtUdPeNL8g6xc1yeSRv76puKCXQFCxvpWfmU3Vgs8sv3TCmKVnjw6WH8ye3ZVrzBaFckvwoBUS335xhTuNRM9xKvVuEH5JVNPCiBinQoAPhhp7oYM97af4PV6oYbsi4qvz9Ln44v7XVUXzJi19H4eoM3vhMBShM5XM8etUxgs5YpZs3haoB59YagEm55zoTwZZ46yM9G4jyKFxoeUfX3Xj5yZGEQ5PmuHidViQG5aETnx7qWjUXtLrmbEZKDoYHiqPg6upRt1D5J2BX5gZmP7cTLosonmnp1qdpBUi9JiLy4N4hHCpYT3W78uBgvN2UWbVdmkDPNWKNP6rpJLJJaytrSie6ABR2Ar4ofsD6bWbGwnMfnYxADh7wYMxQSexsqrymEj6UPXGjKoHHSEbk7x3yyfLT2yS1do5RdiLAxLTJuZyKo9Jtz8m9gXe8MCw5FXeF7QdaXUWyxXkxb4A4uFtCjheQPoNLArzxL24LvYQ7nYnXdpPQKQ6hPeoBFrjBveuttshhDLbZnMK4iQpzEUDfkFd9kEeeSc51YCu3Ny"
	tx, err := ParseTx(rawTx)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	//_ = tx
	txjson, _ := json.Marshal(tx)
	fmt.Println(string(txjson))
}
func Test_ParseUtxo(t *testing.T) {
	id, err := ids.FromString("jnUjZSRt16TcRnZzmh5aMhavwVHz3zBrSN8GfFMTQkzUnoBxC")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%x\n", id.Bytes())
	rawUtxo := "11RyXqXJYZXWYG6Vf8aDn74ZxZNcKzvJ8spETfVoskC8BeGxWJnpDYG6q2j9UVxcwjHYxdJCJtPkW85qr6eLkLcfqvmRERjE9VbjH1wZoXttujtQHxvW3pYdBHSx7tpdkmguSXEzkwhs4R3cRFqMSB2Y1ehw1NdLaBAz2W"
	//rawUtxo := "11FUdfUGKwWgf7Abfhv9BkR96ZugKvpmZ22HczfSwvEEZ4P9R9bfMkbgGSRjvU9skgLtfJ4LEdWifomvBfWuyUpbtW98dimteB5ryjNoFuCxi1P7SUDLiHcq9VmoCB89E4MGEdJM4FYRsMzCJw87EkZCDRNjvW5muEU5Nu"
	fb := formatting.CB58{}
	acodec := spdagvm.Codec{}
	fb.FromString(rawUtxo)
	//fmt.Printf("%x", fb.Bytes)
	//txid := fb.Bytes[:32]
	//index := fb.Bytes[32:36]
	utxo, err := acodec.UnmarshalUTXO(fb.Bytes)
	//if err != nil {
	//	panic(err.Error())
	//}
	utxoJson, _ := json.Marshal(utxo)
	fmt.Println(string(utxoJson))
}
func Test_genAccount2(t *testing.T) {
	return
	fb := formatting.CB58{}

	acodec := spdagvm.Codec{}

	fb.FromString("1111111113QvkUvZtCywinszMHovUGFZtBk5vu2PrunaRWBEVPYM6HLGhNBA2jdAqft91xvKfDbEWiGrxe1BYrjFm4GXkUSo1gme7Xa5A2kMm7vKVTfqRhdWGEcA3TAibg8QB7CrqzvEWz7dLuedJzn72W6Fn3exVRb4qKPPEBQKQEsqbUu1tBQEWpB6QB5JiVBNiW48mYkY1yGkZtSTvBWrcszxe8RhACWpQeCuQ5ZyLaLa7GMDKoJjyWFh8WWCdMhHoR4xGKjqpkakrRmjZtLHy9z7QbdBLW3yp52rck5JqJYMJ4LDCWeiHaZ1vc21gjwx6qJc4imox1H4WYTuEvmeKkf6mVpcWmKifSQNxTp4nzMi2jSKubeDT7LYcfKPpxqUxeaHP2RDv7Ne7qdPzSr5z2xiQ3EDTMyRbLtR8Qum5dzszTabp2DNyiT9b2nS6dXDJUwTBKZLFH92EVaTQggLF5HRrttiF689tt3Uh4CzPaM4mk3VAPeN8xyiWEjX1QHiCNQGv7HYL")
	fmt.Printf("%x\n", fb.Bytes)
	//utxo2 := &avax.UTXO{}
	tx, err := acodec.UnmarshalTx(fb.Bytes)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(fb.String())
	fmt.Println(tx.String())
}
