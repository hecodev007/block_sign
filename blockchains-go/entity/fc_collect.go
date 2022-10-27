package entity

type FcCollect struct {
	Id              int64           `xorm:"'id' pk autoincr"`    //主键自增
	AppId           int64           `xorm:"'app_id'"`            //商户ID
	CreateTime      int64           `xorm:"'create_time'"`       //创建时间
	UpdateTime      int64           `xorm:"'update_time'"`       //创建时间
	CoinName        string          `xorm:"'coin_name'"`         //币种名字
	MainCoinName    string          `xorm:"'main_coin_name'"`    //主链币名 比如 eth的erc20 此处放eth coin_name放代币名
	ContractOrToken string          `xorm:"'contract_or_token'"` //合约或者代币token
	FeeCoinName     string          `xorm:"'fee_coin_name'"`     //手续费币种
	Txid            string          `xorm:"'txid'"`              //交易ID
	FromAddr        string          `xorm:"'from_addr'"`         //发送地址地址
	ChangeAddr      string          `xorm:"'change_addr'"`       //找零地址
	ToAddr          string          `xorm:"'to_addr'"`           //发送地址
	ToAmount        string          `xorm:"'to_amount'"`         //发送金额
	ToFee           string          `xorm:"'to_fee'"`            //发送手续费
	Status          FcCollectStatus `xorm:"'status'"`            //状态
	SendData        string          `xorm:"'send_data'"`         //发送内容体
	ErrData         string          `xorm:"'err_data'"`          //响应错误内容体
	OutOrderNo      string          `xorm:"'out_order_no'"`      //订单号
	Memo            string          `xorm:"'memo'"`              //memo
}

type FcCollectStatus int

//tip:热钱包基本上只有0 2，3状态，冷钱包就需要兼顾所有状态
const (
	FcCollectCreate  FcCollectStatus = 0 //创建完成，等待执行 主要是冷钱包使用
	FcCollectRunning FcCollectStatus = 1 //执行中 主要冷钱包使用
	FcCollectSuccess FcCollectStatus = 2 //执行成功
	FcCollectFail    FcCollectStatus = 3 //执行失败，已废弃
)
