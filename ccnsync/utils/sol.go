package utils

import (
	"github.com/portto/solana-go-sdk/common"
	"log"
)

func FindAssociatedTokenAddress(address, contract string) (string, error) {
	addrPub := common.PublicKeyFromString(address)
	conPub := common.PublicKeyFromString(contract)
	ata, _, err := common.FindAssociatedTokenAddress(addrPub, conPub)
	if err != nil {
		log.Printf("find ata error, err: %v", err)
		return "", err
	}
	return ata.ToBase58(), nil
}
