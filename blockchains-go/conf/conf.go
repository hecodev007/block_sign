package conf

import (
	"github.com/shopspring/decimal"
	"time"
)

type Config struct {
	Env                 string                  `toml:"env"`        // 开发环境
	Mode                string                  `toml:"mode"`       // 运行模式
	MultiLimit          int                     `toml:"multiLimit"` // 多地址出账地址数量限制
	Http                HttpConfig              `toml:"http"`       // http服务器配置
	DB                  DBConfig                `toml:"db"`
	DB2                 DB2Config               `toml:"db2"`
	DBAddrMgr           DBAddrMgr               `toml:"dbAddrMgr"`
	Redis               RedisConfig             `toml:"redis"`
	ClusterRedisConfig  ClusterRedisConfig      `toml:"clusterRedis"`  //集群缓存
	ClusterRedisConfig2 ClusterRedisConfig      `toml:"clusterRedis2"` //集群缓存
	Creates             map[string]CreateConfig `toml:"create"`
	TransferModel       TransferModel           `toml:"transfer"`
	Walletserver        Walletserver            `toml:"walletserver"` // walletserver api服务器
	WalletType          WalletType              `toml:"wallettype"`   // walletserver api服务器
	HotServers          map[string]*HotServers  `toml:"hotservers"`   // 热钱包服务器
	CoinServers         map[string]*CoinServers `toml:"coinservers"`  // 币种服务服务器
	IMBot               IM                      `toml:"im"`           // IM工具消息通知Token
	Encryption          bool                    `toml:"encryption"`   // 加密启用
	Collect             Collect                 `toml:"collect"`      // 归集策略
	UtxoScan            UtxoScanCode            `toml:"scan"`         // utxo二维码扫码限制
	WeChat              WeChat                  `toml:"wechat"`       // 微信告警配置host
	Rate                Rate                    `toml:"rate"`         // utxo费率
	Merge               map[string]*Merge       `toml:"merge"`        // 归集设置
	DataServer          string                  `toml:"dataserver"`   // 补推url
	EthScanCfg          EthScanConfig           `toml:"ethscan"`      // ethscan api
	Other               Other                   `toml:"other"`        // 其他
	CollectCenter       CollectCenter           `toml:"collectCenter"`
	Commandcenter       Commandcenter           `toml:"commandcenter"` //操作中心
}

type HttpConfig struct {
	Port         string        `toml:"port"`         // 服务器端口
	ReadTimeout  time.Duration `toml:"readtimeout"`  // 读取超时,毫秒
	WriteTimeout time.Duration `toml:"writetimeout"` // 写入超时,毫秒
}

type DBConfig struct {
	Master string   `toml:"master"`
	Slaves []string `toml:"slaves"`
}
type DB2Config struct {
	Master string `toml:"master"`
}

type DBAddrMgr struct {
	Master string `toml:"master"`
}

type RedisConfig struct {
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type ClusterRedisConfig struct {
	Cluster bool   `toml:"cluster"`
	Addr    string `toml:"addr"`
	Pwd     string `toml:"pwd"`
}

type CollectConfig struct {
	DingName   string   `toml:"ding_name"`
	DingToken  string   `toml:"ding_token"`
	Mode       string   `toml:"mode"`
	DB         DBConfig `toml:"db"`
	Collectors map[string]*Collect2
}
type Collect2 struct {
	CoinServers
	Name           string   `toml:"name"`
	Spec           string   `toml:"spec"`
	MinAmount      float64  `toml:"min_amount"`
	NeedFee        float64  `toml:"need_fee"`
	AlarmFee       float64  `toml:"alarm_fee"`
	MaxCount       int      `toml:"max_count"`
	IgnoreCoins    []string `toml:"ignore_coins"`
	AssignCoins    []string `toml:"assign_coins"`
	AssignAddress  []string `toml:"assign_address"`
	Mchs           []string `toml:"mchs"`
	Node           string   `toml:"node"`
	UseLatestNonce bool     `toml:"use_latest_nonce"` // 是否使用latest nonce，默认使用pendding nonce
	Nonce          int64    `toml:"nonce"`            // 是否使用这个nonce
	HighGasAddress []string `toml:"highGasAddress"`
}

type CreateConfig struct {
	Interval int    `toml:"interval"`
	Url      string `toml:"url"`
}

type TransferModel struct {
	UtxoModel    []string `toml:"utxo"`
	AccountModel []string `toml:"account"`
}

type Walletserver struct {
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type WalletType struct {
	Cold []string `toml:"cold"`
	Hot  []string `toml:"hot"`
}

type HotServers struct {
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type IM struct {
	DingToken   string `toml:"dingtoken"`   // 钉钉工具token 异常使用
	ReviewToken string `toml:"reviewtoken"` // 钉钉工具token 审核使用 `
	SecretToken string `toml:"secretToken"` // 钉钉机器人带过来的Token
}

type Collect struct {
	Zvc decimal.Decimal `toml:"zvc"`
}

type UtxoScanCode struct {
	Num int `toml:"num"`
}

type CoinServers struct {
	Url      string `toml:"url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type WeChat struct {
	Url string `toml:"url"`
}

type CollectCenter struct {
	Url string `toml:"url"`
}

type Rate struct {
	Ltc   int64 `toml:"ltc"`
	Ghost int64 `toml:"ghost"`
	Bch   int64 `toml:"bch"`
	Zec   int64 `toml:"zec"`
	Avax  int64 `toml:"avax"`
}

type Merge struct {
	BalanceThreshold    decimal.Decimal `toml:"balance_threshold"`     // #合并金额的时候保留阈值金额在冷地址
	MergeThreshold      decimal.Decimal `toml:"merge_threshold"`       // 主链币归集起始金额
	MergeTokenThreshold decimal.Decimal `toml:"merge_token_threshold"` // 代币归集起始金额
}

type EthScanConfig struct {
	Host  string `toml:"host"`
	Token string `toml:"token"`
}

type Other struct {
	XRPSupplementalUrl string `toml:"xrp_supplemental_url"`
}

type Commandcenter struct {
	Url string `toml:"url"`
}
