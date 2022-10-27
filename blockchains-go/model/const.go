package model

//可以移植到环境变量读取,配置文件危险性比较高
const WallectServer = "http://localhost:8080"

type TransferModel string

const TransferModelUtxo TransferModel = "utxo"
const TransferModelAccount TransferModel = "account"
