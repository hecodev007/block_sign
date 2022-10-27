package expand

/*
扩展：解析extrinsic
	substrate2.0的extrinsic都是这样，所以这里的变动其实很小
	这里编写都是为了与github.com/JFJun/substrate-go保持一制，所以会显得有点混乱
*/
import (
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/huandu/xstrings"
	"github.com/rjman-ljm/substrate-go/utils"
)

type ExtrinsicDecoder struct {
	ExtrinsicLength     int              `json:"extrinsic_length"`
	VersionInfo         string           `json:"version_info"`
	ContainsTransaction bool             `json:"contains_transaction"`
	Address             string           `json:"address"`
	Signature           string           `json:"signature"`
	SignatureVersion    int              `json:"signature_version"`
	Nonce               int              `json:"nonce"`
	Era                 string           `json:"era"`
	Tip                 string           `json:"tip"`
	CallIndex           string           `json:"call_index"`
	CallModule          string           `json:"call_module"`
	CallModuleFunction  string           `json:"call_module_function"`
	Params              []ExtrinsicParam `json:"params"`
	me                  *MetadataExpand
	Value               interface{}
}

type ExtrinsicParam struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
	ValueRaw string      `json:"value_raw"`
}

func NewExtrinsicDecoder(meta *types.Metadata) (*ExtrinsicDecoder, error) {
	ed := new(ExtrinsicDecoder)
	var err error
	ed.me, err = NewMetadataExpand(meta)
	if err != nil {
		return nil, err
	}
	return ed, nil
}

func (ed *ExtrinsicDecoder) ProcessExtrinsicDecoder(decoder scale.Decoder, chainName string) error {
	var length types.UCompact
	err := decoder.Decode(&length)
	if err != nil {
		return fmt.Errorf("decode extrinsic: length error: %v\n", err)
	}
	ed.ExtrinsicLength = int(utils.UCompactToBigInt(length).Int64())
	vi, err := decoder.ReadOneByte()
	if err != nil {
		return fmt.Errorf("decode extrinsic: read version info error: %v\n", err)
	}
	ed.VersionInfo = utils.BytesToHex([]byte{vi})
	ed.ContainsTransaction = utils.U256(ed.VersionInfo).Int64() >= 80
	//大多数都是84了，所以只处理84
	if ed.VersionInfo == "04" || ed.VersionInfo == "84" {
		if ed.ContainsTransaction {
			// 1. 解析from地址
			var address MultiAddress
			err = decoder.Decode(&address)
			if err != nil {
				return fmt.Errorf("decode extrinsic: decode address error: %v\n", err)
			}
			ed.Address = utils.BytesToHex(address.AccountId[:])
			//2。解析签名版本
			var sv types.U8
			err = decoder.Decode(&sv)
			if err != nil {
				return fmt.Errorf("decode extrinsic: decode signature version error: %v\n", err)
			}
			ed.SignatureVersion = int(sv)
			// 3。 解析签名
			if ed.SignatureVersion == 2 {
				//解析 ecdsa signature
				sig := make([]byte, 65)
				err = decoder.Read(sig)
				if err != nil {
					return fmt.Errorf("decode extrinsic: decode ecdsa signature error: %v\n", err)
				}
				ed.Signature = utils.BytesToHex(sig)
			} else {
				// 解析 sr25519 signature
				var sig types.Signature
				err = decoder.Decode(&sig)
				if err != nil {
					return fmt.Errorf("decode extrinsic: decode sr25519 signature error: %v\n", err)
				}
				ed.Signature = sig.Hex()
			}
			// 4. 解析era
			var era types.ExtrinsicEra
			err = decoder.Decode(&era)
			if err != nil {
				return fmt.Errorf("decode extrinsic: decode era error: %v\n", err)
			}
			if era.IsMortalEra {
				eraBytes := []byte{era.AsMortalEra.First, era.AsMortalEra.Second}
				ed.Era = utils.BytesToHex(eraBytes)
			}
			//5. 解析nonce
			var nonce types.UCompact
			err = decoder.Decode(&nonce)
			if err != nil {
				return fmt.Errorf("decode extrinsic: decode nonce error: %v\n", err)
			}
			//new

			ed.Nonce = int(utils.UCompactToBigInt(nonce).Int64())
			// 6.解析tip
			var tip types.UCompact

			err = decoder.Decode(&tip)
			if err != nil {
				return fmt.Errorf("decode tip error: %v\n", err)
			}
			ed.Tip = fmt.Sprintf("%d", utils.UCompactToBigInt(tip).Int64())
		}
		//处理callIndex
		callIndex := make([]byte, 2)
		err = decoder.Read(callIndex)
		if err != nil {
			return fmt.Errorf("decode extrinsic: read call index bytes error: %v\n", err)
		}
		ed.CallIndex = xstrings.RightJustify(utils.IntToHex(callIndex[0]), 2, "0") +
			xstrings.RightJustify(utils.IntToHex(callIndex[1]), 2, "0")
	} else {
		return fmt.Errorf("extrinsics version %s is not support", ed.VersionInfo)
	}
	if ed.CallIndex != "" {
		err = ed.decodeCallIndex(decoder, chainName)
		if err != nil {
			return fmt.Errorf("decodeCallIndex err: %v\n", err)
		}
	}
	result := map[string]interface{}{
		"extrinsic_length": ed.ExtrinsicLength,
		"version_info":     ed.VersionInfo,
	}
	if ed.ContainsTransaction {
		result["account_id"] = ed.Address
		result["signature"] = ed.Signature
		result["nonce"] = ed.Nonce
		result["era"] = ed.Era
	}
	if ed.CallIndex != "" {
		result["call_code"] = ed.CallIndex
		result["call_module_function"] = ed.CallModuleFunction
		result["call_module"] = ed.CallModule
	}
	result["nonce"] = ed.Nonce
	result["era"] = ed.Era
	result["tip"] = ed.Tip
	result["params"] = ed.Params
	result["length"] = ed.ExtrinsicLength
	ed.Value = result
	return nil
}

func (ed *ExtrinsicDecoder) decodeCallIndex(decoder scale.Decoder, chainName string) error {
	var err error
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			err = fmt.Errorf("decode call catch panic ,err=%v\n", errs)
		}
	}()
	//	解析 call index
	// 这里只能硬编码了，因为decode函数的原因，无法动态的根据type name去解析
	// 这里我只解析自己想要的，比如说Timestamp,Balance.transfer,Utility.batch
	modName, callName, err := ed.me.MV.FindNameByCallIndex(ed.CallIndex)
	if err != nil {
		return fmt.Errorf("decode call: %v\n", err)
	}
	ed.CallModule = modName
	ed.CallModuleFunction = callName
	switch modName {
	case "System":
		if callName == "remark" {
			var s string
			err = decoder.Decode(&s)
			if err != nil {
				return fmt.Errorf("decode call: decode Timestamp.set error: %v\n", err)
			}

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:  "remark",
					Type:  "string",
					Value: s,
				})
		}

	case "Timestamp":
		if callName == "set" {
			//Compact<Moment>
			var u types.UCompact
			err = decoder.Decode(&u)
			if err != nil {
				return fmt.Errorf("decode call: decode Timestamp.set error: %v\n", err)
			}

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:  "now",
					Type:  "Compact<Moment>",
					Value: utils.UCompactToBigInt(u).Int64(),
				})
		}
	case "Balances":
		if callName == "transfer" || callName == "transfer_keep_alive" {
			// 0 ---> 	Address
			var addrValue string
			var address MultiAddress
			err = decoder.Decode(&address)
			if err != nil {
				return fmt.Errorf("decode call: decode Balances.transfer.Address error: %v\n", err)
			}
			addrValue = utils.BytesToHex(address.AccountId[:])

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:     "dest",
					Type:     "Address",
					Value:    addrValue,
					ValueRaw: addrValue,
				})
			// 1 ----> Compact<Balance>
			var b types.UCompact
			err = decoder.Decode(&b)
			if err != nil {
				return fmt.Errorf("decode call: decode Balances.transfer.Compact<Balance> error: %v\n", err)
			}

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:  "value",
					Type:  "Compact<Balance>",
					Value: utils.UCompactToBigInt(b).Int64(),
				})
		}
	case "Multisig":
		if ed.checkChainX(chainName) && chainName == ClientNameChainXAsset {
			/// Chain is ChainX-XBTC
			if callName == "as_multi" {
				//1. decode threshold
				var threshold uint16
				err = decoder.Decode(&threshold)
				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "threshold",
						Type:  "uint16",
						Value: threshold,
					})

				//2. decode otherSignatories
				var otherSignatories []string
				var address []types.AccountID
				//var bt byte
				//err = decoder.Decode(&bt)
				err = decoder.Decode(&address)
				if err != nil {
					return fmt.Errorf("decode call: decode Multi.as_multi.OtherSignatories error: %v\n", err)
				}
				for _, add := range address {
					otherSignatories = append(otherSignatories, utils.BytesToHex(add[:]))
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "other_signatories",
						Type:  "vec<AccountId>",
						Value: otherSignatories,
					})

				//3. docode TimePoint
				var tp []uint32
				var option []byte

				//tp := TimePointSafe32{}
				//err = decoder.Decode(&option)
				//fmt.Printf("err is %v\n", err)

				var hasValue bool
				_ = decoder.DecodeOption(&hasValue, option)
				//fmt.Printf("timePoint is %v\n", err)

				if hasValue {
					var height uint32
					var index uint32

					err = decoder.Decode(&height)
					err = decoder.Decode(&index)

					//blockNumber := types.NewOptionU32(height)

					//if !option {
					//	fmt.Errorf("TimePoint.Height is Not Safe!")
					//}

					tp = append(tp, height)
					tp = append(tp, index)
				} else {
					tp = nil
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "maybe_timepoint",
						Type:  "TimePointSafe32",
						Value: tp,
					})

				//4. decode call => transfer
				vec := new(Vec)
				var tc xTransferOpaqueCall
				//err = vec.ProcessFirstVec(decoder, tc)
				err = vec.ProcessOpaqueCallVec(decoder, tc)
				if err != nil {
					return fmt.Errorf("decode call: ulti => Assets.transfer error: %v, chain is %v\n", err, chainName)
				}

				ep := ExtrinsicParam{}
				ep.Name = "calls"
				ep.Type = "Vec<Call>"
				var result []interface{}

				for _, value := range vec.Value {
					tcv := value.(*xTransferOpaqueCall)
					btCallIdx, err := ed.me.MV.GetCallIndex("XAssets", "transfer")
					if err != nil {
						return fmt.Errorf("decode ChainX => Multisig.as_multi: get  Multisig.as_multi call index error: %v, chain is %v\n", err, chainName)
					}

					/// Check for XAssetsTransfer
					data := tcv.Value.(map[string]interface{})
					if data["call_index"].(string) == "0300" || data["call_index"].(string) == "0000" || data["call_index"].(string) == "03ff" || data["call_index"].(string) == "00ff" {
						/// Polkadot is 0503, substrate is 0603, ChainX is 0603
						data["call_index"] = btCallIdx
					}
					callIndex := data["call_index"].(string)

					if err != nil {
						return fmt.Errorf("decode Multisig.as_multi: get  Balances.transfer_keep_alive call index error: %v, chain is %v\n", err, chainName)
					}
					if callIndex == btCallIdx {
						mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
						if err != nil {
							return fmt.Errorf("decode Multisig.as_multi: get call index error: %v\n", err)
						}
						if mn != "XAssets" {
							return fmt.Errorf("decode Utility.batch:  call module name is not 'XAssets' ,NAME=%s", mn)
						}
						data["call_function"] = cn
						data["call_module"] = mn
						result = append(result, data)
					}
					ep.Value = result
					ed.Params = append(ed.Params, ep)
				}
				//5. decode store_call
				var storeCall bool
				err = decoder.Decode(&storeCall)
				if err != nil {
					log.Debug("decode call: decode Multi.as_multi.store_call error", "Error", err, "Chain", chainName)
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "store_call",
						Type:  "bool",
						Value: storeCall,
					})

				//6. decode Weight
				var maxWeight types.Weight
				err = decoder.Decode(&maxWeight)
				if err != nil {
					log.Debug("decode call: decode Multi.as_multi.max_weight error", "Error", err, "Chain", chainName)
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "max_weight",
						Type:  "uint64",
						Value: maxWeight,
					})
			}
		} else {
			if callName == "as_multi" {
				//1. decode threshold
				var threshold uint16
				err = decoder.Decode(&threshold)
				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "threshold",
						Type:  "uint16",
						Value: threshold,
					})

				//2. decode otherSignatories
				var otherSignatories []string
				var address []types.AccountID
				//var bt byte
				//err = decoder.Decode(&bt)
				err = decoder.Decode(&address)
				if err != nil {
					return fmt.Errorf("decode call: decode Multi.as_multi.OtherSignatories error: %v, chain is %v\n", err, chainName)
				}
				for _, add := range address {
					otherSignatories = append(otherSignatories, utils.BytesToHex(add[:]))
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "other_signatories",
						Type:  "vec<AccountId>",
						Value: otherSignatories,
					})

				//3. docode TimePoint
				var tp []uint32
				var option []byte

				//tp := TimePointSafe32{}
				//err = decoder.Decode(&option)
				//fmt.Printf("err is %v\n", err)

				var hasValue bool
				_ = decoder.DecodeOption(&hasValue, option)
				//fmt.Printf("timePoint is %v\n", err)

				if hasValue {
					var height uint32
					var index uint32

					err = decoder.Decode(&height)
					err = decoder.Decode(&index)

					//blockNumber := types.NewOptionU32(height)

					//if !option {
					//	fmt.Errorf("TimePoint.Height is Not Safe!")
					//}

					tp = append(tp, height)
					tp = append(tp, index)
				} else {
					tp = nil
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "maybe_timepoint",
						Type:  "TimePointSafe32",
						Value: tp,
					})

				//4. decode call => transfer
				vec := new(Vec)
				var tc TransferOpaqueCall
				//err = vec.ProcessFirstVec(decoder, tc)
				err = vec.ProcessOpaqueCallVec(decoder, tc)
				if err != nil {
					return fmt.Errorf("decode call: decode Utility.batch => Balances.transfer error: %v, chain is %v\n", err, chainName)
				}

				ep := ExtrinsicParam{}
				ep.Name = "calls"
				ep.Type = "Vec<Call>"
				var result []interface{}

				for _, value := range vec.Value {
					tcv := value.(*TransferOpaqueCall)
					btCallIdx, err := ed.me.MV.GetCallIndex("Balances", "transfer")
					if err != nil {
						return fmt.Errorf("decode Multisig.as_multi: get  Multisig.as_multi call index error: %v, chain is %v\n", err, chainName)
					}
					btkaCallIdx, err := ed.me.MV.GetCallIndex("Balances", "transfer_keep_alive")

					/// Check for BalanceTransfer
					data := tcv.Value.(map[string]interface{})
					if data["call_index"].(string) == "0300" || data["call_index"].(string) == "0000" || data["call_index"].(string) == "03ff" {
						/// Polkadot is 0503, substrate is 0603, ChainX is 0603
						data["call_index"] = btCallIdx
					}
					callIndex := data["call_index"].(string)

					if err != nil {
						return fmt.Errorf("decode Multisig.as_multi: get  Balances.transfer_keep_alive call index error: %v, chain is %v\n", err, chainName)
					}
					if callIndex == btCallIdx || callIndex == btkaCallIdx {
						mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
						if err != nil {
							return fmt.Errorf("decode Multisig.as_multi: get call index error: %v, chain is %v\n", err, chainName)
						}
						if mn != "Balances" {
							return fmt.Errorf("decode Utility.batch:  call module name is not 'Balances' ,NAME=%s", mn)
						}
						data["call_function"] = cn
						data["call_module"] = mn
						result = append(result, data)
					}
					ep.Value = result
					ed.Params = append(ed.Params, ep)
				}
				//5. decode store_call
				var storeCall bool
				err = decoder.Decode(&storeCall)
				if err != nil {
					log.Debug("decode call: decode Multi.as_multi.store_call error", "Error", err, "Chain", chainName)
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "store_call",
						Type:  "bool",
						Value: storeCall,
					})

				//6. decode Weight
				var maxWeight uint64
				err = decoder.Decode(&maxWeight)
				if err != nil {
					log.Debug("decode call: decode Multi.as_multi.max_weight error", "Error", err, "Chain", chainName)
				}

				ed.Params = append(ed.Params,
					ExtrinsicParam{
						Name:  "max_weight",
						Type:  "uint64",
						Value: maxWeight,
					})
			}
		}
	case "Utility":
		if callName == "batch_all" {
			callName = "batch"
		}
		/// Check whether ChainX
		if ed.checkChainX(chainName) && chainName == ClientNameChainXAsset {
			/// Chain is ChainX-XBTC
			if callName == "batch" || callName == "batch_all" {
				vec := new(Vec)

				// 1: XAssets.Transfer
				var tc xTransferCall
				err = vec.ProcessFirstVec(decoder, tc)
				if err != nil {
					return fmt.Errorf("decode call: decode Utility.batch => XAssets.transfer error: %v\n", err)
				}

				// 2: System.remark
				var rc RemarkCall
				err := vec.ProcessSecondVec(decoder, rc)
				if err != nil {
					return fmt.Errorf("decode call: decode Utility.batch => System.remark error: %v\n", err)
				}

				//utils.CheckStructData(vec.Value)
				ep := ExtrinsicParam{}
				ep.Name = "calls"
				ep.Type = "Vec<Call>"
				var result []interface{}

				for i, value := range vec.Value {
					if i == 0 {
						tcv := value.(*xTransferCall)
						//检查一下是否为xAssetsTransfer
						data := tcv.Value.(map[string]interface{})
						callIndex := data["call_index"].(string)
						btCallIdx, err := ed.me.MV.GetCallIndex("XAssets", "transfer")
						if err != nil {
							return fmt.Errorf("decode Utility.batch: get  XAssets.transfer call index error: %v\n", err)
						}
						if callIndex == btCallIdx {
							mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
							if err != nil {
								return fmt.Errorf("decode Utility.batch: get call index error: %v\n", err)
							}
							if mn != "XAssets" {
								return fmt.Errorf("decode Utility.batch:  call module name is not 'Balances' ,NAME=%s\n", mn)
							}
							data["call_function"] = cn
							data["call_module"] = mn
							result = append(result, data)
						} else {
							return fmt.Errorf("decode Utility.batch error: not XAssets.transfer, %v\n", err)
						}
					}
					if i == 1 {
						tcv := value.(*RemarkCall)
						//检查一下是否为System.remark
						data := tcv.Value.(map[string]interface{})
						callIndex := data["call_index"].(string)
						srCallIdx, err := ed.me.MV.GetCallIndex("System", "remark")
						if err != nil {
							return fmt.Errorf("decode Utility.batch: get  Balances.transfer call index error: %v\n", err)
						}
						if callIndex == srCallIdx {
							mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
							if err != nil {
								return fmt.Errorf("decode Utility.batch: get call index error: %v\n", err)
							}
							if mn != "System" {
								return fmt.Errorf("decode Utility.batch:  call module name is not 'Balances' ,NAME=%s\n", mn)
							}
							data["call_function"] = cn
							data["call_module"] = mn
							result = append(result, data)
						} else {
							return fmt.Errorf("decode Utility.batch error: not System.remark, %v\n", err)
						}
					}
				}
				ep.Value = result
				ed.Params = append(ed.Params, ep)
			}
		} else {
			if callName == "batch" || callName == "batch_all" {
				vec := new(Vec)
				//BEGIN: Custom decode
				// 1: Balances.Transfer
				// 0--> calls   Vec<Call>
				var tc TransferCall
				err = vec.ProcessFirstVec(decoder, tc)
				if err != nil {
					return fmt.Errorf("decode call: decode Utility.batch => Balances.transfer error: %v\n", err)
				}

				// 2: System.remark
				var rc RemarkCall
				err := vec.ProcessSecondVec(decoder, rc)
				if err != nil {
					return fmt.Errorf("decode call: decode Utility.batch => System.remark error: %v\n", err)
				}

				//utils.CheckStructData(vec.Value)
				ep := ExtrinsicParam{}
				ep.Name = "calls"
				ep.Type = "Vec<Call>"
				var result []interface{}

				for i, value := range vec.Value {
					if i == 0 {
						tcv := value.(*TransferCall)
						//检查一下是否为BalanceTransfer
						data := tcv.Value.(map[string]interface{})
						callIndex := data["call_index"].(string)
						btCallIdx, err := ed.me.MV.GetCallIndex("Balances", "transfer")
						if err != nil {
							return fmt.Errorf("decode Utility.batch: get  Balances.transfer call index error: %v\n", err)
						}
						btkaCallIdx, err := ed.me.MV.GetCallIndex("Balances", "transfer_keep_alive")
						if err != nil {
							return fmt.Errorf("decode Utility.batch: get  Balances.transfer_keep_alive call index error: %v\n", err)
						}
						if callIndex == btCallIdx || callIndex == btkaCallIdx {
							mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
							if err != nil {
								return fmt.Errorf("decode Utility.batch: get call index error: %v\n", err)
							}
							if mn != "Balances" {
								return fmt.Errorf("decode Utility.batch:  call module name is not 'Balances' ,NAME=%s\n", mn)
							}
							data["call_function"] = cn
							data["call_module"] = mn
							result = append(result, data)
						} else {
							return fmt.Errorf("decode Utility.batch error: not Balances.transfer, %v\n", err)
						}
					}
					if i == 1 {
						tcv := value.(*RemarkCall)
						//检查一下是否为System.remark
						data := tcv.Value.(map[string]interface{})
						callIndex := data["call_index"].(string)
						srCallIdx, err := ed.me.MV.GetCallIndex("System", "remark")
						if err != nil {
							return fmt.Errorf("decode Utility.batch: get  Balances.transfer call index error: %v\n", err)
						}
						if callIndex == srCallIdx {
							mn, cn, err := ed.me.MV.FindNameByCallIndex(callIndex)
							if err != nil {
								return fmt.Errorf("decode Utility.batch: get call index error: %v\n", err)
							}
							if mn != "System" {
								return fmt.Errorf("decode Utility.batch:  call module name is not 'Balances' ,NAME=%s\n", mn)
							}
							data["call_function"] = cn
							data["call_module"] = mn
							result = append(result, data)
						} else {
							return fmt.Errorf("decode Utility.batch error: not System.remark, %v\n", err)
						}
					}
				}
				ep.Value = result
				ed.Params = append(ed.Params, ep)
			}
		}
	case "XAssets":
		if callName == "transfer" {
			var address MultiAddress
			var addrValue string

			// 0 ---> 	Address
			err = decoder.Decode(&address)
			if err != nil {
				return fmt.Errorf("decode call: decode Balances.transfer.Address error: %v\n", err)
			}
			addrValue = utils.BytesToHex(address.AccountId[:])

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:     "dest",
					Type:     "Address",
					Value:    addrValue,
					ValueRaw: addrValue,
				})

			// 1 ----> Compact<AssetId>
			var id types.UCompact
			err = decoder.Decode(&id)
			if err != nil {
				return fmt.Errorf("decode call: decode Balances.transfer.Address error: %v\n", err)
			}
			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:  "id",
					Type:  "Compact<AssetId>",
					Value: types.U32(utils.UCompactToBigInt(id).Int64()),
				})

			// 2 ----> Compact<Balance>
			var b types.UCompact
			err = decoder.Decode(&b)
			if err != nil {
				return fmt.Errorf("decode call: decode Balances.transfer.Compact<Balance> error: %v\n", err)
			}

			ed.Params = append(ed.Params,
				ExtrinsicParam{
					Name:  "value",
					Type:  "Compact<Balance>",
					Value: utils.UCompactToBigInt(b).Int64(),
				})
		}
	default:
		// unsopport
		return nil

	}
	return nil
}

func (ed *ExtrinsicDecoder) checkChainX(chainName string) bool {
	var isChainX = false
	_, err := ed.me.MV.GetCallIndex("XAssets", "transfer")
	if err == nil {
		isChainX = true
	}
	if isChainX || chainName == ClientNameChainX  || chainName == ClientNameChainXAsset {
		return true
	} else {
		return false
	}
}
