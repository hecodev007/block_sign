# web3.go-thk
#Installation

go get

go get -u github.com/thinkey-dev/web3.go

#Usage
通过交易hash获取交易详情(web3.thk.GetTxByHash)
```
var response =  web3.thk. GetTransactionByHash("2","0x3cbd7226fb9d4c9bbd27cdc230a647ecd19aa2997e23ab899778026093f45326")
```
```
response:
{
"Transaction": {
"chainID": 2,
"from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
"input": "0x",
"nonce": 42,
"to": "0x6ea0fefc17c877c7a4b0f139728ed39dc134a967",
"value": 2333
},
"blockHeight": 117354,
"contractAddress": "0x0000000000000000000000000000000000000000",
"logs": null,
"out": "0x",
"root": null,
"status": 1,
"transactionHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```
# Thinkey Web3.go SDK接口文档

# 1. 获取账户余额(web3.thk.GetBalance)


请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| address | string | true | 账户地址 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| address | string | true | 账户地址 |
| nonce | int | true | 交易的发起者在之前进行过的交易数量 |  
| balance| bigint | true | 账户余额 |  
| storageRoot| string | false | 合约存储数据的hash(没有合约返回null) | 
|codeHash| string | false | 合约代码的hash(没有合约返回null) | 

请求示例:
```
var response =  web3.thk. GetAccount("2","0x2c7536e3605d9c16a7a3d7b1898e529396a65c23")
```
```
{
"address": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
"balance": 9.99999985e+26,
"codeHash": null,
"nonce": 43
"storageRoot": null
}
```

# 2. 执行一笔交易(web3.thk.SendTX)

请求参数：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| Transaction | dict | true | 交易详情 |
Transaction：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| fromChainId | string | true | 交易发起账户地址的链id |
| toChainId | string | true | 交易接受账户地址的链id |
|from | string | true | 交易发起账户地址 |
|to | string | true | 交易接受账户地址 |
| nonce | string | true | 交易的发起者在之前进行过的交易数量 |   
| value | string | true | 转账金额 |  
| input | string | true | 调用合约时的参数 |  
| sig | string | true | 交易签名 |  
| pub | string | true | 公钥 |  

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| TXhash | string | true | 交易hash |


请求示例:
```
transaction := util.Transaction{
   ChainId: "2", FromChainId: "2", ToChainId: "3", From: 0x2c7536e3605d9c16a7a3d7b1898e529396a65c23,
   To: 0x0000000000000000000000000000000000020000
 Value: "0", Input: 0x000000022c7536e3605d9c16a7a3d7b1898e529396a65c23000000000000000a000000034fa1c4e6182b6b7f3bca273390cf587b50b4731100000000000456440101, Nonce: 10),
}
privatekey, err := crypto.HexToECDSA(key)
err = connection.Thk.SignTransaction(&transaction, privatekey)
txhash, err := connection.Thk.SendTx(&transaction)
```
```
response:
{
"TXhash": "0x22024c2e429196ac76d0e557ac0cf6141f5b500c56fde845582b837c9dab236b"
}
```

# 3. 通过交易hash获取交易详情(web3.thk.GetTxByHash)

请求参数：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
|hash | string | true | 交易hash |


响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| Transaction | dict | true | 交易详情 |
| root | string | true | 保存了创建该receipt对象时，“账户”的当时状态 |
| status | int | true | 交易状态: 1:成功, 0:失败 |
| logs | array[dict] | false | 这个交易产生的日志对象数组 |
| transactionHash | string | true | 交易hash |
| contractAddress | string | true | 合约账户地址 |
| out | string | true | 调用返回结果数据 |

Transaction:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainID | int| true | 链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | string | true | 交易的发起者在之前进行过的交易数量 |
| val | string | true | 转账金额 |
| input | string | true | 调用合约时的参数 |


请求示例:
```
var response =  web3.thk. GetTransactionByHash("2","0x3cbd7226fb9d4c9bbd27cdc230a647ecd19aa2997e23ab899778026093f45326")
```
```
response:
{
"Transaction": {
"chainID": 2,
"from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
"input": "0x",
"nonce": 42,
"to": "0x6ea0fefc17c877c7a4b0f139728ed39dc134a967",
"value": 2333
},
"blockHeight": 117354,
"contractAddress": "0x0000000000000000000000000000000000000000",
"logs": null,
"out": "0x",
"root": null,
"status": 1,
"transactionHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
```

# 4. 获取链信息(web3.thk.GetStats)

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |


响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| currentheight |bigint | true |  当前块高 |
| txcount | int | true | 总交易数 |
| tps | int | true | 每秒交易数 |
| tpsLastEpoch | int | true | 上一时期交易数 |
| lives | int | true | 链的已存活时间 |
| accountcount | int | true | 账户数 |
| epochlength | int | true | 当前时期包含多少块 |
| epochduration | int | true | 当前时期运行时间 |
| lastepochduration | int | true | 上一时期的运行时间 |
| currentcomm | array | true | 当前这条链的委员会成员 |

请求示例:
```
var response =  web3.thk.GetStats("1")
```
```
{
"accountcount": 0,
"currentcomm": [
 "0xd1f889690f8c75bbada89a4c8893b8bf6fe29be3b5c3d8a2d772024a340d59d375f39ed88498666a57da10af885ad63a414f8a10153fb739eb1ebfcef57cc883", "0xe90a151759bf070969aae664e00502bb08568c85a73874492a3ec480c5178d5da29c790896fc62106e32d172819dec94202ff90f3b7ba3e6adf38508bc58cf43",
 "0x84385cc16d8e0a47909ee998d51370e5f56d7c85716e045c99760bedb180346da7d00b575ba23b76ffcd0969ae84e1e6b6943ec408f40b44825128577d8a895d",
 "0xd0c7107542af7e0019e1340a77a00131d60f49f5543de76b1d5768660e6d694b5dee3e206049bf0009d2859db0b7378240667d85eeb8138426efe9fd3568ebe3"
],
"currentheight": 124262,
"epochduration": 797,
"epochlength": 300,
"lastepochduration": 796,
"lives": 242529,
"tps": 0,
"tpsLastEpoch": 0,
"txcount": 10
}
```

# 5. 获取指定账户在对应链上一定高度范围内的交易信息(web3.thk.GetTransactions)

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| address | string | true | 账户地址 |
| startHeight | string | true | 查询的起始块高 |
| endHeight | string | true | 查询的截止块高 |

响应参数：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| [] | []transactons | true | 交易信息数组 |


transactons：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | int | true | 链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | int | true | 交易的发起者在之前进行过的交易数量 |
| value | int | true | 转账金额 |
| timestamp | int | true | 交易的时间戳 |
| input | string | true | 调用合约时的参数 |
| hash | string | true | 交易hash |


请求示例:
```
var response = web3.thk. GetTransactions(“2”,”0x2c7536e3605d9c16a7a3d7b1898e529396a65c23”,”50”,”100”)
```
```
Response:
[
    {
        "chainId": 2,
        "from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
        "to": "0x0000000000000000000000000000000000020000",
        "nonce": 0,
        "value": 0,
        "input": "0x000000022c7536e3605d9c16a7a3d7b1898e529396a65c230000000000000000000000034fa1c4e6182b6b7f3bca273390cf587b50b4731100000000000456440101",
        "hash": "0x0ea5dad47833fc6286357b6bd6c1a4e910def5f4432a1a59bde0f816c3dd18e0",
        "timestamp": 1560425588
    },
    {
        "chainId": 2,
        "from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
        "to": "0x133c5bfef5d486052b061b44af113f20057341a8",
        "nonce": 1,
        "value": 0,
        "input": "0xa9059cbb00000000000000000000000066261e3faf00ef1537b22f37d8db85f57066f58f0000000000000000000000000000000000000000000000000000000000004e20",
        "hash": "0x1dbbda2d229db82ff12b3bea82d49225e6bebd645def4c06da157ddbe5660066",
        "timestamp": 1560425596
    }
]
```

# 6. 调用交易（web3.thk.CallTransaction）

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| Transaction | dict | true | 交易详情 |
请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| fromChainId | string | true | 交易发起账户地址的链id |
| toChainId | string | true | 交易接受账户地址的链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | string | true | 交易的发起者在之前进行过的交易数量 |
| value | string | true | 转账金额 |
| input | string | true | 调用合约时的参数 |




响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| Transaction | dict | true | 交易详情 |
| root | string | true | 保存了创建该receipt对象时，“账户”的当时状态 |
| status | int | true | 交易状态: 1:成功, 0:失败 |
| logs | array[dict] | true | 这个交易产生的日志对象数组 |
| transactionHash | string | true | 交易hash |
| contractAddress | string | true | 合约账户地址 |
| out | string | true | 调用返回结果数据 |


Transaction:

| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | int | true | 链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | int | true | 交易的发起者在之前进行过的交易数量 |
| value | int | true | 转账金额 |
| input | string | true | 调用合约时的参数 |



请求示例:
```
 var connection = web3.NewWeb3(providers.NewHTTPProvider("192.168.1.13:8089", 10, false))
from := "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23"
to := "0x6ea0fefc17c877c7a4b0f139728ed39dc134a967"
transaction := util.Transaction{
   ChainId: "2", FromChainId: "2", ToChainId: "2", From: from,
   To: to, Value: "2333", Input: "", Nonce: "1",
}
res, err := connection.Thk.CallTransaction(&transaction)
```
```
Response:
{
    {2 0x2c7536e3605d9c16a7a3d7b1898e529396a65c23 0x6ea0fefc17c877c7a4b0f139728ed39dc134a967 11 2333 0x}  1  0xa7540a40565982de81fa4261b689b556e6a28a6de7d86e4428f429b8259d86ae 0x0000000000000000000000000000000000000000 0x 0 }
```

# 7. 获取指定块高信息(web3.thk.GetBlockHeader)

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| height | string | true | 查询块的块高 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| hash | string | true | 此块的hash |
| previoushash | string | true | 父块的hash |
| chainid | int | true | 链id |
| height | int | true | 查询块的块高 |
| mergeroot | string | true | 合并其他链转账数据hash |
| deltaroot | string | true | 跨链转账数据hash |
| stateroot | string | true | 状态hash |
| txcount | int | true | 交易总数 |
| timestamp | int | true | 时间戳 |

请求示例:
```
var response =  web3.thk. GetBlockHeader("1","30")
```
```
response:
{
"chainid": 1,
"deltaroot": null,
"hash": "0x6bd6a3d1068a3b748edc7ef70aee98749e33ddc3e03e10ca49dc4ca5fad4237c",
"height": 30,
"mergeroot": null,
"previoushash":"0xfed93048d70bd961582bbd0498f109b78d471a90a3e7b1e0aa65b91f4982de97",
stateroot":"0xacf1890a60e805815cbf6e93fdb9f7a0184bc51290a39802e0c67e961ab41f35",
"timestamp": 1560425446,
"txcount": 0
}
```

# 8. 获取指定块的交易(web3.thk.getBlockTxs)


请求参数
       
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| height | string | true | 查询块的块高 |
| page | string | true | 页码 |
| size | string | true | 页的大小 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| elections | dict | true | 交易详情 |
| accountchanges | array | true | 交易信息 |

accountchanges:
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainid | string | true | 链id |
| height | int| true | 查询的起始块高 |
| from | string| true | 交易发起账户地址 |
| to | string| true | 交易接受账户地址 |
| nonce | int| true | 交易的发起者在之前进行过的交易数量 |
| value | int| true | 转账金额 |
| timestamp | int| true | 交易的时间戳 |


请求示例:
```
var response =  web3.thk. GetBlockTxs("2","433","1","10")
```
```
response:
{
"elections": null,
"accountchanges": [
{
"chainid": 2,
"from": "0x2c7536e3605d9c16a7a3d7b1898e529396a65c23",
"to": "0x66261e3faf00ef1537b22f37d8db85f57066f58f",
"nonce": 11,
"input": "0x",
"hash":"0xf976d11b1a1593e242d5ed8ad77cf6df1b12c5a9c3e50d4c98fbf2ace7738bb3",
"value": 10000000000000000000,
"timestamp": 1560426508
}
]
}
```

# 9. 编译合约(web3.thk.compileContract)
  
请求参数：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| contract | string | true | 合约代码 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| code | string | true | 编译后的合约代码 |
| info | dict | true | 合约信息 |


请求示例:
```
Var response = web3.thk.compileContract('2', 'pragma solidity >= 0.4.0;contract test {uint storedData; function set(uint x) public { storedData = x;} function get() public view returns (uint) { return storedData;}}');
```
```
response:
{
"test": {
"code": "0x608060405234801561001057600080fd5b5060be8061001f6000396000f3fe6080604052348015600f57600080fd5b5060043610604e577c0100000000000000000000000000000000000000000000000000000000600035046360fe47b1811460535780636d4ce63c14606f575b600080fd5b606d60048036036020811015606757600080fd5b50356087565b005b6075608c565b60408051918252519081900360200190f35b600055565b6000549056fea165627a7a723058205f13a3c1870823036833f92b7ac23a38f4bb1d9b737c36f0ea70ded514af2a6c0029",
"info": {
"source": "pragma solidity >= 0.4.0;contract test {uint storedData; function set(uint x) public { storedData = x;} function get() public view returns (uint) { return storedData;}}",
"language": "Solidity",
"languageVersion": "0.5.2",
"compilerVersion": "0.5.2",
"compilerOptions": "--combined-json bin,abi,userdoc,devdoc,metadata --optimize",
"abiDefinition": [
{
"constant": false,
"inputs": [
{
"name": "x",
"type": "uint256"
}
],
"name": "set",
"outputs": [],
"payable": false,
"stateMutability": "nonpayable",
"type": "function"
},
{
"constant": true,
"inputs": [],
"name": "get",
"outputs": [
{
"name": "",
"type": "uint256"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
}
],
"userDoc": {
"methods": {}
},
"developerDoc": {
"methods": {}
},
"metadata": "{\"compiler\":{\"version\":\"0.5.2+commit.1df8f40c\"},\"language\":\"Solidity\",\"output\":{\"abi\":[{\"constant\":false,\"inputs\":[{\"name\":\"x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}],\"devdoc\":{\"methods\":{}},\"userdoc\":{\"methods\":{}}},\"settings\":{\"compilationTarget\":{\"<stdin>\":\"test\"},\"evmVersion\":\"byzantium\",\"libraries\":{},\"optimizer\":{\"enabled\":true,\"runs\":200},\"remappings\":[]},\"sources\":{\"<stdin>\":{\"keccak256\":\"0xa906fc7673818a545ec91bd707cee4d4549c5bf8ae684ddfcee70b0417fd07df\",\"urls\":[\"bzzr://fa15822d315f9ccb19e8e3f658c14e843995e90d9ba8a142ed144edd26eb017b\"]}},\"version\":1}"
}
}
}
```

# 10.web3.thk.Ping

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| address | string | true | ip+端口 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| nodeId | string | true | 节点id |
| version | string | true | 版本 |
| isDataNode | bool | true | 是否是数据节点 |
| dataNodeOf | int | true | 数据节点 |
| lastMsgTime | int64 | true | 上一个信息时间 |
| lastEventTime | int64 | true | 上一个事件时间 |
| lastBlockTime | int64 | true | 上一个块时间 |
| overflow | bool | true | 溢出 |
| lastBlocks |map | true | 最后一个块 |
| opTypes | map | true | 类型 |


请求示例:
```
var response =  web3.thk.Ping("192.168.1.13:22010")
```
```
Response:
{
    "nodeId": "0x5e17128ba224a96d6e84be0c7f899febea26c55c78940610d78a0d22dbd0ab03cc3233491de0b5eb770dbf850b509bd191723df4fc40520bcbab565d46543d6e",
    "version": "V1.0.0",
    "isDataNode": true,
    "dataNodeOf": 0,
    "lastMsgTime": 1560850367,
    "lastEventTime": 1560850367,
    "lastBlockTime": 1560850367,
    "overflow": false,
    "lastBlocks": {
        "0": 159927
    },
    "opTypes": {
        "0": [
            "DATA"
        ]
    }
}
```

# 11. 生成支票的证明(web3.thk.RpcMakeVccProof)

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| transaction | dict | true | 交易对象 |
  
transaction：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| fromChainId | string | true | 交易发起账户地址的链id |
| toChainId | string | true | 交易接受账户地址的链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | string | true | 交易的发起者在之前进行过的交易数量 |
| value | string | true | 转账金额 |
| ExpireHeight | int | true | 过期高度 |

响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| input | string | true | 生成的支票证明 |


请求示例:
```
var response =  web3.thk. RpcMakeVccProof("2","0x2c7536e3605d9c16a7a3d7b1898e529396a65c23","0x4fa1c4e6182b6b7f3bca273390cf587b50b47311","2","1","284228","25")
```
```
response:
{
"input": "0x95000000022c7536e3605d9c16a7a3d7b1898e529396a65c230000000000000019000000034fa1c4e6182b6b7f3bca273390cf587b50b473110000000000045644010102a301e64dc0d4e0daf294ed06960285719c81945b516931f46f689b7c330a041445e9d0c23c93941093a1b0dfbdbf5e039a614e6fc5e077e373c8c706fbd529454ee64e9dcae974df7b346bc200008080940a934080c2ffff8081000462f62879bcb53487b2b5a7705622002ceef2792208cd5596957e787d413679bc9ed0f9274d52040f0c7edc5465031d95141fb5e170c213ca2dade245d87fb18782552b2a7176b962e3a53b772c88ecdd99ccd8dc677b32f08394be0c72ad602d70eeb30cf600eb18284aef075aebac26863d38b639d7859ff5058266ed6fb72000010a9424930080c20000c0ed5a50458bbdee9150090681de9f958784b4de973a05869ac006cdfe62f9bbaa810005bdf41875ebf61043535eb71e9ae6e1409200a44f84f6c8e364a20999ef58a02ab485c3b70ab1171549b8ba7d7e7b2bd734563318bea9782b5328a53bb429421fde7c23dfe73d9fb6cc75caa077409f6c47c06425e617441fbd788634617136a0eca078605c1b0ad6ff4323f7c23307585d3dddd504f96e7a7f722f9802d2a1b7d9333ab2116cd47b78b4a9df5c24a62ec1a559d90d92476e2f9ab4fb6c536194000110"
}
}
```

# 12. 生成取消支票的证明(web3.thk.MakeCCCExistenceProof)

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| transaction | dict | true | 交易对象 |
  
transaction：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| fromChainId | string | true | 交易发起账户地址的链id |
| toChainId | string | true | 交易接受账户地址的链id |
| from | string | true | 交易发起账户地址 |
| to | string | true | 交易接受账户地址 |
| nonce | string | true | 交易的发起者在之前进行过的交易数量 |
| value | string | true | 转账金额 |
| ExpireHeight | int | true | 过期高度 |


响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| input | string | true | 调用合约时的参数 |
| existence | bool | true | 是否存过支票 |


请求示例:
```
var response =  web3.thk. MakeCCCExistenceProof("3","0x2c7536e3605d9c16a7a3d7b1898e529396a65c23","0x4fa1c4e6182b6b7f3bca273390cf587b50b47311","3","3","1","284228","66")
```
```
response:
{
"existence": false,
"input": "0x96000000032c7536e3605d9c16a7a3d7b1898e529396a65c230000000000000042000000034fa1c4e6182b6b7f3bca273390cf587b50b473110000000000045644010103a301e6cfc0948eaf91f6454fc73d796a5d219080ed568159a17915972046471ebddef4839a9294a1fe93a1b0dfc4e4f4c830bbe175349a2e40f2b36e9ff2c1882c072a50d1173744662e6a57c20000c040c4e4f4c830bbe175349a2e40f2b36e9ff2c1882c072a50d1173744662e6a57809404934080c2fdfd808100044d398e46e89ea357f971c9c164d7410d8d320866c225d2d7acc558bdfffa13c7b0d714935f98aaf7965de3caa4b1ea069c944ddcde81dcd35718b765a3ace55e87ff331742bc9bd34e940861649e30a570e4a3695507dd97cbe6353b5b4dca61c8d3a1f913d6feecc9432edbb0d3e141054c6b5a4b4f7a58c34c3de7e68f8ad6000103919425930080c20000c038da7c56bfaab9b73df7714a931570683b63b746790a04886b5cf339535f0a9481000516652bf7d0a127262a7e44d4846584069c4a6ca866c8b5d85785f01b52de4590edec92d65fa4c0c82a6351785089d275c865ba3ae403548671a8c862bdf8ae6ba7aefc67d042b85d3ddf744d02bb7fa1427ad5375b0b5acd84305dd3ffaed995eca078605c1b0ad6ff4323f7c23307585d3dddd504f96e7a7f722f9802d2a1b7fdf0b76142223be82c013d51c23f864ba5a7e07461b66ee053bc52b3e5bda584000111"
}
```

# 13.获取链结构（web3.thk.GetChainInfo）

请求参数
        
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainIds | []int | true | 链id（备注：传空代表所有） |


响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| [] | []chainInfo | true | 链信息数组 |

chainInfo：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | int | true | 链id |
| datanodes | []dataNode | true | 数据节点群 |
| mode | int | true | 模式 |
| parent | int | true | 父 |

dataNode：
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| dataNodeId | int | true | 数据节点id |
| dataNodeIp | string | true | 数据节点ip |
| dataNodePort | int | true | 数据节点端口 |

请求示例:
```
var response =  web3.thk.GetChainInfo([])
```
```
response:
[
{
"chainId": 0,
"datanodes": [
{
"dataNodeId": "0x5e17128ba224a96d6e84be0c7f899febea26c55c78940610d78a0d22dbd0ab03cc3233491de0b5eb770dbf850b509bd191723df4fc40520bcbab565d46543d6e",
"dataNodeIp": "192.168.1.13",
"dataNodePort": 22010
}
],
"mode": 5,
"parent": 1048576
},
{
"chainId": 1,
"datanodes": [
{
"dataNodeId": "0x96dc94580e0eadd78691807f6eac9759b9964daa8b46da4378902b040e0eb102cb48413308d2131e9e5557321f30ba9287794f689854e6d2e63928a082e79286",
"dataNodeIp": "192.168.1.13",
"dataNodePort": 22014
}
],
"mode": 6,
"parent": 0
},
{
"chainId": 2,
"datanodes": [
{
"dataNodeId": "0xa93b150f11c422d8700554859281be8e34a91a859e0e021af186002c7e4a2661ea2467a63b417030d68e2fdddeb4342943dff13225da77124abf912fd092f71f",
"dataNodeIp": "192.168.1.13",
"dataNodePort": 22018
}
],
"mode": 6,
"parent": 0
},
{
"chainId": 3,
"datanodes": [
{
"dataNodeId": "0x783f4b2490461ecfd8ee8d3451e434de06bacb0ffff56de53a33fe545589094fa0b929eeaa62dc5203d1e831ccdd37d206d0b85b193921efb223bf0cb2f37b4c",
"dataNodeIp": "192.168.1.13",
"dataNodePort": 22022
}
],
"mode": 7,
"parent": 1
},
{
"chainId": 4,
"datanodes": [
{
"dataNodeId": "0x44c98ab831f3ca4553e491bba06753e959ceb55d43e18bc76539572feb1e0dbaf2fbfc19f571d6544e82be1c7c39760f6a023d4be4dcb9473dd580c731d03926",
"dataNodeIp": "192.168.1.13",
"dataNodePort": 22026
}
],
"mode": 7,
"parent": 1
}
]
```

# 14.获取委员会详情（web3.thk.GetCommittee）

请求参数
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | string | true | 链id |
| epoch | string | true | 参选轮次 |


响应参数:
           
| 参数名 | 类型 | 是否必须| 含义 |
| :------:| :------: | :------: | :------: |
| chainId | int | true | 链id |
| MemberDetails | []string | true | 委员会详情 |
| Epoch | int | true | 参选轮次 |


请求示例:
```
var response =  web3.thk.GetCommittee("1","411")
```
```
response:
[
    "0xd1f889690f8c75bbada89a4c8893b8bf6fe29be3b5c3d8a2d772024a340d59d375f39ed88498666a57da10af885ad63a414f8a10153fb739eb1ebfcef57cc883",
    "0xe90a151759bf070969aae664e00502bb08568c85a73874492a3ec480c5178d5da29c790896fc62106e32d172819dec94202ff90f3b7ba3e6adf38508bc58cf43",
    "0x84385cc16d8e0a47909ee998d51370e5f56d7c85716e045c99760bedb180346da7d00b575ba23b76ffcd0969ae84e1e6b6943ec408f40b44825128577d8a895d",
    "0xd0c7107542af7e0019e1340a77a00131d60f49f5543de76b1d5768660e6d694b5dee3e206049bf0009d2859db0b7378240667d85eeb8138426efe9fd3568ebe3"
]
```
