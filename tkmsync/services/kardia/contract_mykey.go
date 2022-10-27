package kardia

import (
	"bytes"
	"context"
	"errors"
	"math"
	"math/big"
	"strings"

	// token "./erc20" // for demo

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// LogTransfer ..
type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

// LogApproval ..
type LogApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
}

var (
	client *ethclient.Client

	//ethChainHost = "https://eth.mykey.tech"
	//ethChainHost = "http://13.114.44.225:31545"

	transferLogicAddress      = common.HexToAddress("0x1c2349acbb7f83d07577692c75b6d7654899bf10")
	transferLogicEnteredTopic = common.HexToHash("0x3efc190d59645f005a5974aa84aa94401ad997938870e7b2aa74a45138ad679b") //
	transferLogicABI          = "[{\"constant\":true,\"inputs\":[{\"name\":\"_key\",\"type\":\"address\"}],\"name\":\"getKeyNonce\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"accountStorage\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_accountStorage\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"account\",\"type\":\"address\"}],\"name\":\"TransferLogicInitialised\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"bytes\"},{\"indexed\":true,\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"TransferLogicEntered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"LogicInitialised\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"_account\",\"type\":\"address\"}],\"name\":\"initAccount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\"},{\"name\":\"_signature\",\"type\":\"bytes\"},{\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"enter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferEth\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferErc20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_approvedSpender\",\"type\":\"address\"},{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_token\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferApprovedErc20\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_nftContract\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"},{\"name\":\"_safe\",\"type\":\"bool\"}],\"name\":\"transferNft\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_approvedSpender\",\"type\":\"address\"},{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_nftContract\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"},{\"name\":\"_safe\",\"type\":\"bool\"}],\"name\":\"transferApprovedNft\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\"},{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"}]"
	transferLogicAbi, _       = abi.JSON(strings.NewReader(string(transferLogicABI)))

	opCodeTransferEth   = crypto.Keccak256Hash([]byte("transferEth(address,address,uint256)")).Bytes()[:4]
	opCodeTransferErc20 = crypto.Keccak256Hash([]byte("transferErc20(address,address,address,uint256)")).Bytes()[:4]
)

func InitEthClient(ethChainHost string) error {
	var err error
	client, err = ethclient.Dial(ethChainHost)
	if err != nil {
		return err
	}

	return nil
}

type transferLogicEnteredData struct {
	Data []byte
}

func MyKeyProcessTransferLogic(vLog types.Log) (sender, receiver, amountFloatStr, txid string, err error) {
	var data transferLogicEnteredData

	transferLogicAbi.Unpack(&data, "TransferLogicEntered", vLog.Data)
	// log.Println(hex.EncodeToString(data.Data))
	eventData := data.Data[4:]
	if bytes.Compare(data.Data[:4], opCodeTransferEth) == 0 {
		quantityExact := big.NewInt(0).SetBytes(data.Data[32*2+4:])
		sender = common.BytesToAddress(eventData[12:32]).Hex()
		receiver = common.BytesToAddress(eventData[32+12 : 32*2]).Hex()
		//symbol := "ETH"
		ethValue := new(big.Float).Quo(big.NewFloat(0).SetInt(quantityExact), big.NewFloat(math.Pow10(18)))
		amountFloatStr = ethValue.String()
		//fmt.Println(symbol, sender, receiver, ethValue.String(), vLog.TxHash.Hex())
		txid = vLog.TxHash.Hex()
		return sender, receiver, amountFloatStr, txid, nil
	}
	return "", "", "", "", errors.New("opCodeTransferEth error")
}

func GetMKTransEvent(height int64) (sender, receiver, amountFloatStr, txid string, err error) {
	//header, err := client.HeaderByNumber(context.Background(), nil)
	//if err != nil {
	//	return err
	//}
	//log.Println("latest block:", header.Number.String())
	startBlock := height
	loopEndBlock := height
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(startBlock),
		ToBlock:   big.NewInt(loopEndBlock),
		Addresses: []common.Address{
			transferLogicAddress,
		},
		Topics: [][]common.Hash{{transferLogicEnteredTopic}},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return "", "", "", "", err
	}
	for _, vLog := range logs {
		// log.Printf("%d %d\n", vLog.BlockNumber, vLog.Index)
		// b, _ := json.Marshal(vLog)
		// log.Println(string(b))
		if vLog.Topics[0] == transferLogicEnteredTopic {
			//理论上有多个，但是目前暂时发现只有一个
			sender, receiver, amountFloatStr, txid, err = MyKeyProcessTransferLogic(vLog)
			return "", "", "", "", nil
		}
	}

	return "", "", "", "", errors.New("empty mykey trans")
}
