package sol

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"solsign/common/conf"
	"solsign/common/log"
	"solsign/common/validator"
	"time"

	"github.com/portto/solana-go-sdk/assotokenprog"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/shopspring/decimal"

	"github.com/portto/solana-go-sdk/tokenprog"

	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/sysprog"

	"github.com/portto/solana-go-sdk/types"
)

func BuildTx(params *validator.TelosSignParams, getPri func(string, string) ([]byte, error)) (rawTx []byte, err error) {

	if params.ContractAddress == "" {
		//主链币转账
		return BuildSysTransfer(params, getPri)
	} else {
		//合约转账
		return BuildTokenTransfer(params, getPri)
	}
}

func BuildSysTransfer(params *validator.TelosSignParams, getPri func(string, string) ([]byte, error)) (rawTx []byte, err error) {
	if params.ContractAddress != "" {
		return rawTx, errors.New("签名服务内部错误,合约地址不为空")
	}
	frompri, err := getPri(params.MchName, params.FromAddress)
	if err != nil {
		return
	}
	log.Info(hex.EncodeToString(frompri))
	fromSigner := types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(frompri))
	feeSigner := fromSigner
	//log.Info(fromSigner.PublicKey.ToBase58(), params.FromAddress)
	//return nil, errors.New("test")
	var signers []types.Account
	signers = append(signers, fromSigner)
	if params.FeeAddress == "" {
		//params.FeeAddress = params.FromAddress
	} else {
		feepri, err := getPri(params.MchName, params.FeeAddress)
		if err != nil {
			return rawTx, err
		}
		feeSigner = types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(feepri))
		signers = append(signers, feeSigner)
	}
	rawtx, err := types.CreateRawTransaction(types.CreateRawTransactionParam{
		Instructions: []types.Instruction{
			sysprog.Transfer(
				common.PublicKeyFromString(params.FromAddress), // from
				common.PublicKeyFromString(params.ToAddress),   // to
				params.Amount.BigInt().Uint64(),                // 1 SOL
			),
		},
		Signers:         signers,
		FeePayer:        feeSigner.PublicKey,
		RecentBlockHash: params.BlockHash,
	})

	return rawtx, err
}

func BuildTokenTransfer(params *validator.TelosSignParams, getPri func(string, string) ([]byte, error)) (rawTx []byte, err error) {
	if params.ContractAddress == "" {
		return rawTx, errors.New("合约地址不能为空")
	}
	if params.Amount.Cmp(decimal.NewFromInt(math.MaxInt64)) > 0 {
		return rawTx, errors.New("额度超过签名能支持的上限")
	}
	client := NewClient(conf.GetConfig().Node.Url)
	_, toTokenAddress, decimals, err := client.BalanceOf(params.ToAddress, params.ContractAddress)
	if err != nil {
		return
	}
	//没有子地址则创建
	if toTokenAddress == "" {
		txid, err := CreateTokenAccount(params, getPri)
		if err != nil {
			return rawTx, err
		}
		log.Info(params.OrderId, "创建子地址交易:"+txid)
	}
	st := time.Now()
BalanceOf:
	_, toTokenAddress, decimals, err = client.BalanceOf(params.ToAddress, params.ContractAddress)
	if err != nil {
		log.Info(err.Error())
	}
	//等待十秒交易上链
	if toTokenAddress == "" && time.Since(st) < time.Second*60 {
		goto BalanceOf
	}
	if toTokenAddress == "" {
		return nil, errors.New("创建子地址失败")
	}
	log.Info("合约子地址: ", toTokenAddress)
	//block, err := client.GetRecentBlockhash()
	//if err != nil {
	//	log.Info(params.OrderId, err.Error())
	//	return
	//}
	//params.BlockHash = block.Blockhash

	_, fromTokenAddress, _, err := client.BalanceOf(params.FromAddress, params.ContractAddress)
	if err != nil {
		return
	}
	//合约转账
	frompri, err := getPri(params.MchName, params.FromAddress)
	if err != nil {
		return
	}
	fromSigner := types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(frompri))
	var signers []types.Account
	feeSigner := fromSigner
	signers = append(signers, fromSigner)
	if params.FeeAddress == "" {
		//params.FeeAddress = params.FromAddress
	} else {
		feepri, err := getPri(params.MchName, params.FeeAddress)
		if err != nil {
			return rawTx, err
		}
		feeSigner = types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(feepri))
		signers = append(signers, feeSigner)
	}
	mintPubkey := common.PublicKeyFromString(params.ContractAddress)
	toTokenPubkey := common.PublicKeyFromString(toTokenAddress)
	fromTokenPubkey := common.PublicKeyFromString(fromTokenAddress)
	var signerPubkeys []common.PublicKey
	for _, v := range signers {
		signerPubkeys = append(signerPubkeys, v.PublicKey)
	}
	rawtx, err := types.CreateRawTransaction(types.CreateRawTransactionParam{
		Instructions: []types.Instruction{
			tokenprog.TransferChecked(
				fromTokenPubkey,
				toTokenPubkey,
				mintPubkey,
				fromSigner.PublicKey,
				signerPubkeys,
				params.Amount.BigInt().Uint64(),
				uint8(decimals),
			),
		},
		Signers:         signers,
		FeePayer:        feeSigner.PublicKey,
		RecentBlockHash: params.BlockHash,
	})

	return rawtx, err
}

func CreateTokenAccount(params *validator.TelosSignParams, getPri func(string, string) ([]byte, error)) (txid string, err error) {
	if params.ContractAddress == "" {
		return txid, errors.New("合约地址为空")
	}
	frompri, err := getPri(params.MchName, params.FromAddress)
	if err != nil {
		return
	}
	fromSigner := types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(frompri))
	var signers []types.Account
	signers = append(signers, fromSigner)
	feeSigner := fromSigner
	if params.FeeAddress == "" {
		//params.FeeAddress = params.FromAddress
	} else {
		feepri, err := getPri(params.MchName, params.FeeAddress)
		if err != nil {
			return txid, err
		}
		feeSigner = types.AccountFromPrivateKeyBytes(ed25519.NewKeyFromSeed(feepri))
		signers = append(signers, feeSigner)
	}
	toPubkey := common.PublicKeyFromString(params.ToAddress)
	tokenMint := common.PublicKeyFromString(params.ContractAddress)
	rawtx, err := types.CreateRawTransaction(types.CreateRawTransactionParam{
		Instructions: []types.Instruction{
			assotokenprog.CreateAssociatedTokenAccount(
				fromSigner.PublicKey,
				toPubkey,
				tokenMint,
			),
		},
		Signers:         signers,
		FeePayer:        feeSigner.PublicKey,
		RecentBlockHash: params.BlockHash,
	})
	client := NewClient(conf.GetConfig().Node.Url)
	txid, err = client.SendRawTransaction(rawtx)
	return txid, err
}
