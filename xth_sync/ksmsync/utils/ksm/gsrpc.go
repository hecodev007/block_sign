package ksm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	scalecodec "github.com/itering/scale.go"
	"github.com/itering/scale.go/source"
	scaletypes "github.com/itering/scale.go/types"
	"github.com/itering/scale.go/utiles"
	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
	gsrc "github.com/yanyushr/go-substrate-rpc-client/v3"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
	"golang.org/x/crypto/blake2b"
)

var (
	KsmMetadataDecoder scalecodec.MetadataDecoder
)

func init() {

	KsmMetadataDecoder = scalecodec.MetadataDecoder{}

	KsmMetadataDecoder.Init(utiles.HexToBytes(kusamaV14))
	_ = KsmMetadataDecoder.Process()

	ksmFile, err := ioutil.ReadFile(fmt.Sprintf("%s.json", "network/kusama"))
	if err != nil {
		panic(err)
	}

	scaletypes.RegCustomTypes(source.LoadTypeRegistry(ksmFile))
}

type Client struct {
	Api *gsrc.SubstrateAPI

	Meta               *types.Metadata
	ChainName          string //链名字
	SpecVersion        int
	TransactionVersion int
	genesisHash        types.Hash
	Url                string
}

func NewClient(url string) (*Client, error) {
	c := new(Client)
	c.Url = url
	var err error

	// 初始化rpc客户端
	c.Api, err = gsrc.NewSubstrateAPI(url)
	if err != nil {
		return nil, err
	}
	//检查当前链运行的版本
	err = c.checkRuntimeVersion()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c Client) GetAccountInfo(address string, meta *types.Metadata) (types.AccountInfo, error) {
	var accountInfo types.AccountInfo

	pubKey := GetPublicFromAddr(address, KSMPrefix)

	key, err := types.CreateStorageKey(meta, "System", "Account", pubKey, nil)
	if err != nil {
		return accountInfo, err
	}

	_, err = c.Api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		return accountInfo, err
	}
	return accountInfo, nil
}

func (c Client) GetGenesisHash() types.Hash {
	if c.genesisHash.Hex() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
		return c.genesisHash
	}
	hash, err := c.Api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return types.Hash{}
	}
	c.genesisHash = hash
	return hash
}

func (c Client) reConnect() error {
	api, err := gsrc.NewSubstrateAPI(c.Url)
	if err != nil {
		return err
	}
	c.Api = api
	return nil
}

func (c Client) checkRuntimeVersion() error {
	v, err := c.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		if !strings.Contains(err.Error(), "tls: use of closed connection") {
			return fmt.Errorf("init runtime version error,err=%v", err)
		}
		//	重连处理，这是因为第三方包的问题，所以只能这样处理了了
		err := c.reConnect()
		if err != nil {
			return fmt.Errorf("reconnect error: %v", err)
		}

		v, err = c.Api.RPC.State.GetRuntimeVersionLatest()
		if err != nil {
			return fmt.Errorf("init runtime version error,aleady reconnect,err: %v", err)
		}
	}
	c.TransactionVersion = int(v.TransactionVersion)
	c.ChainName = v.SpecName
	specVersion := int(v.SpecVersion)
	//检查metadata数据是否有升级
	if specVersion != c.SpecVersion {
		c.Meta, err = c.Api.RPC.State.GetMetadataLatest()
		if err != nil {
			return fmt.Errorf("init metadata error: %v", err)
		}
		c.SpecVersion = specVersion
	}
	return nil
}

func (c Client) SetGenesisHash(hash types.Hash) {
	c.genesisHash = hash
}

//获取RPC服务URL
func (c Client) URL() string {
	return c.Url
}

type HeadResponse struct {
	Number decimal.Decimal `json:"number"`
	Hash   string          `json:"hash"`
}

func (c Client) GetBestHeight() (int64, error) {
	return c.GetBlockCount()
}
func (c Client) GetBlockCount() (bestBlockCount int64, err error) {
	header, err := c.Api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		return 0, err
	}

	return int64(header.Number), nil
}

func (c Client) GetBlock(h int64) (*BlockResponse, error) {
	hash, err := c.Api.RPC.Chain.GetBlockHash(uint64(h))
	if err != nil {
		return nil, err
	}

	block, err := c.Api.RPC.Chain.GetBlock(hash)
	if err != nil {
		return nil, err
	}

	ret := BlockResponse{}

	ret.ParentHash = block.Block.Header.ParentHash.Hex()
	ret.BlockHash = hash.Hex()
	ret.Height = int64(block.Block.Header.Number)

	//
	meta, err := c.Api.RPC.State.GetMetadata(hash)
	if err != nil {
		return nil, err
	}
	err = c.ParseExtrinsic(meta, block.Block.Header.ParentHash, block.Block.Extrinsics, &ret)
	if err != nil {
		return nil, err
	}

	//get eventrecords
	er, err := c.GetEventRecords(hash)
	if err != nil {
		return nil, err
	}
	//
	err = c.ParseEventForExtrinsic(er, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c Client) ParseExtrinsic(meta *types.Metadata, blockHash types.Hash, extrinsics []types.Extrinsic, rsp *BlockResponse) error {

	transIndex, err := meta.FindCallIndex("Balances.transfer")
	if err != nil {
		return err
	}

	transLiveIndex, err := meta.FindCallIndex("Balances.transfer_keep_alive")
	if err != nil {
		return err
	}

	for i, v := range extrinsics {
		var ext Extrinsic
		ext.ExtrinsicIndex = i

		if transIndex.SectionIndex == v.Method.CallIndex.SectionIndex && transIndex.MethodIndex == v.Method.CallIndex.MethodIndex {
			//transfer hash
			rawtx, err := types.EncodeToBytes(v)
			if err != nil {
				return err
			}

			h := blake2b.Sum256(rawtx)
			rawtx = h[:]

			ext.Txid = types.HexEncodeToString(rawtx)

			// create address
			from, err := CreateAddress(v.Signature.Signer.AsID[:], KSMPrefix)
			if err != nil {
				return err
			}
			ext.FromAddress = from

			args := Args{}
			err = types.DecodeFromBytes(v.Method.Args, &args)
			if err != nil {
				return err
			}
			to, err := CreateAddress(args.To.AsID[:], KSMPrefix)
			if err != nil {
				return err
			}
			ext.ToAddress = to

			ext.Signature = v.Signature.Signature.AsSr25519.Hex()
			extStr, err := types.EncodeToHexString(v)
			if err != nil {
				return err
			}

			ext.Fee, err = c.PartialFee(extStr, blockHash.Hex())
			if err != nil {
				return err
			}

			rsp.Extrinsics = append(rsp.Extrinsics, &ext)

		} else if transLiveIndex.SectionIndex == v.Method.CallIndex.SectionIndex && transLiveIndex.MethodIndex == v.Method.CallIndex.MethodIndex {
			//transfer hash
			rawtx, err := types.EncodeToBytes(v)
			if err != nil {
				return err
			}
			h := blake2b.Sum256(rawtx)
			rawtx = h[:]

			ext.Txid = types.HexEncodeToString(rawtx)

			// create address
			from, err := CreateAddress(v.Signature.Signer.AsID[:], KSMPrefix)
			if err != nil {
				return err
			}
			ext.FromAddress = from

			args := Args{}
			err = types.DecodeFromBytes(v.Method.Args, &args)
			if err != nil {
				return err
			}
			to, err := CreateAddress(args.To.AsID[:], KSMPrefix)
			if err != nil {
				return err
			}
			ext.ToAddress = to

			ext.Signature = v.Signature.Signature.AsSr25519.Hex()
			extStr, err := types.EncodeToHexString(v)
			if err != nil {
				return err
			}

			ext.Fee, err = c.PartialFee(extStr, blockHash.Hex())
			if err != nil {
				return err
			}

			rsp.Extrinsics = append(rsp.Extrinsics, &ext)
		}
	}
	return nil
}

func (c Client) ParseEventForExtrinsic(er *EventModelRecord, ret *BlockResponse) error {
	res := make([]*EventResult, 0)

	if len(er.GetBalancesTransfer()) > 0 {
		//failed event
		failedMap := make(map[int]bool)
		for _, failed := range er.GetSystemExtrinsicFailed() {
			extrinsicIdx := failed.GetExtrinsicIdx()
			failedMap[int(extrinsicIdx)] = true

		}

		for _, ebt := range er.GetBalancesTransfer() {
			// if !ebt.Phase.IsApplyExtrinsic {
			// 	continue
			// }
			extrinsicIdx := ebt.GetExtrinsicIdx()
			var r EventResult
			r.ExtrinsicIdx = extrinsicIdx
			//
			eventParams := ebt.Params.([]scalecodec.EventParam)
			if len(eventParams) == 3 {
				fr, _ := types.HexDecodeString(eventParams[0].Value.(string))
				from, err := CreateAddress(fr, KSMPrefix)
				if err != nil {
					return err
				}
				r.From = from

				t, _ := types.HexDecodeString(eventParams[1].Value.(string))
				to, err := CreateAddress(t, KSMPrefix)
				if err != nil {
					return err
				}
				r.To = to

				r.Amount = eventParams[2].Value.(string)
				//r.Weight = c.getWeight(&events, r.ExtrinsicIdx)
				res = append(res, &r)
			}

		}
		//
		for _, e := range ret.Extrinsics {
			e.Status = "fail"
			e.Type = "transfer"

			if len(res) > 0 {
				for _, r := range res {
					if e.ExtrinsicIndex == r.ExtrinsicIdx {
						if e.ToAddress == r.To {
							if failedMap[e.ExtrinsicIndex] {
								e.Status = "fail"
							} else {
								e.Status = "success"
							}
							e.Type = "transfer"
							e.Amount = r.Amount
							e.ToAddress = r.To
							// log.Info("transfer:", e)
							// log.Info("event:", r)
						}
					}
				}
			}
		}

	}
	return nil
}

func (c Client) GetBlockByNum(h int64) (ret *BlockResponse, err error) {
	return c.GetBlock(h)
}

func (c Client) GetExtrinsicsByNum(height int64) (Extrinsics []string, err error) {

	var hash string
	err = c.Api.Client.Call(&hash, "chain_getBlockHash", height)
	if err != nil {
		return nil, err
	}

	block := types.SignedBlock{}
	err = c.Api.Client.Call(&block, "chain_getBlock", hash)
	if err != nil {
		return
	}

	extrinsics := make([]string, 0)
	for _, v := range block.Block.Extrinsics {
		extr, err := types.EncodeToHexString(v)
		if err != nil {
			log.Info(err)
			continue
		}
		extrinsics = append(extrinsics, extr)
	}

	return extrinsics, nil
}

func (c Client) GetEventRecords(blockHash types.Hash) (*EventModelRecord, error) {
	meta, err := c.Api.RPC.State.GetMetadata(blockHash)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil)
	if err != nil {
		log.Info(err)
		return nil, err
	}

	raw, err := c.Api.RPC.State.GetStorageRaw(key, blockHash)
	if err != nil {
		log.Info(key, blockHash, err)
		return nil, err
	}

	/*************scale.go************/
	eventDecoder := scalecodec.EventsDecoder{}
	option := scaletypes.ScaleDecoderOption{Metadata: &KsmMetadataDecoder.Metadata}
	eventDecoder.Init(scaletypes.ScaleBytes{Data: *raw}, &option)
	eventDecoder.Process()

	v := eventDecoder.Value.([]interface{})
	evr := make(EventModelRecord, 0)
	for _, vv := range v {
		vvv := vv.(map[string]interface{})
		ev := EventModel{
			Phase:        vvv["phase"],
			ExtrinsicIdx: vvv["extrinsic_idx"],
			Type:         vvv["type"],
			ModuleId:     vvv["module_id"],
			EventId:      vvv["event_id"],
			Params:       vvv["params"],
			Topics:       vvv["topics"],
		}
		evr = append(evr, ev)
	}
	return &evr, nil
	/************************************/

	// events := DotEventRecords{}
	// err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
	// return &events, err
}

func (c Client) PartialFee(rawtx string, blockhash string) (fee string, err error) {
	result := new(QueryInfo)
retry:
	err = c.Api.Client.Call(result, "payment_queryInfo", rawtx, blockhash)
	if err != nil {
		time.Sleep(10 * time.Second)
		goto retry
	}

	return result.PartialFee.String(), nil
}

func ToString(i interface{}) string {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	default:
		b, _ := json.Marshal(i)
		val = string(b)
	}
	return val
}
