package waves

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"wavessign/common/log"
	"wavessign/common/validator"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
	"github.com/wavesplatform/gowaves/pkg/settings"
	"golang.org/x/net/context"
)

func GentAccount() (addr string, pri string, err error) {
	var scheme = settings.MainNetSettings.AddressSchemeCharacter
	var seed [32]byte
	_, err = io.ReadFull(rand.Reader, seed[:])
	if err != nil {
		return "", "", err
	}
	_, pubkey, err := crypto.GenerateKeyPair(seed[:])
	if err != nil {
		return "", "", err
	}
	address, err := proto.NewAddressFromPublicKey(scheme, pubkey)
	if err != nil {
		return "", "", err
	}
	addr = address.String()
	pri = hex.EncodeToString(seed[:])
	return
}
func VerifyAddress(address string) error {
	addr, err := proto.NewAddressFromString(address)
	if err != nil {
		return err
	}
	//proto.addressVersion
	addressVersion := uint8(0x01)
	if addr[0] != addressVersion {
		return errors.New("错误的地址版本")
	}
	if addr[1] != settings.MainNetSettings.AddressSchemeCharacter {
		return errors.New("错误的地址主题")
	}
	return nil
}
func GetBalance(address, contractAddr string) (value uint64, valuestr string, err error) {
	cli, err := client.NewClient()
	if err != nil {
		return 0, "", err
	}
	addr, err := proto.NewAddressFromString(address)
	if err != nil {
		return 0, "", err
	}
	if contractAddr == "" {
		balance, _, err := cli.Addresses.Balance(context.Background(), addr)
		if err != nil {
			return 0, "", err
		}
		return balance.Balance, strconv.FormatUint(balance.Balance, 10), nil
	}
	digest, err := crypto.NewDigestFromBase58(contractAddr)
	if err != nil {
		return 0, "", err
	}

	balance, _, err := cli.Assets.BalanceByAddressAndAsset(context.Background(), addr, digest)
	if err != nil {
		return 0, "", err
	}
	return balance.Balance, strconv.FormatUint(balance.Balance, 10), nil

}
func Sign(params *validator.SignParams, pri string) (txid string, tx *proto.TransferWithSig, err error) {
	pribytes, err := hex.DecodeString(pri)
	if err != nil {
		return "", nil, err
	}
	sk, pk, err := crypto.GenerateKeyPair(pribytes)
	if err != nil {
		return "", nil, err
	}
	toAddr, err := proto.NewAddressFromString(params.ToAddress)
	if err != nil {
		return "", nil, err
	}
	//waves := proto.NewOptionalAssetWaves()
	amountAsset := proto.NewOptionalAssetWaves()
	feeAsset := proto.NewOptionalAssetWaves()
	if params.ContractAddress != "" {
		Assert, err := proto.NewOptionalAssetFromString(params.ContractAddress)
		if err != nil {
			return "", nil, err
		}
		amountAsset = *Assert
	}
	tx = proto.NewUnsignedTransferWithSig(pk, amountAsset, feeAsset, params.Timestamp, params.Value.BigInt().Uint64(), params.Fee, proto.NewRecipientFromAddress(toAddr), []byte("attachment"))
	err = tx.Sign(settings.MainNetSettings.AddressSchemeCharacter, sk)
	txid = tx.ID.String()
	tx.GetID(settings.MainNetSettings.AddressSchemeCharacter)
	return
}
func SendRawTransaction(tx proto.Transaction) error {
	cli, err := client.NewClient()
	if err != nil {
		return err
	}
	resp, err := cli.Transactions.Broadcast(context.Background(), tx)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Info(string(body))
	return nil
}
