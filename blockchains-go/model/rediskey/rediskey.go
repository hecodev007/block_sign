package rediskey

//redis key值定义

const (
	BTC_RATE       = "btc_rate"        // btc费率
	BTC_UTXO_LOCK  = "btc_utxo_lock_"  // 临时锁定utxo,两分钟释放,k = txid v = outorderid
	USDT_UTXO_LOCK = "usdt_utxo_lock_" // 临时锁定utxo,两分钟释放,k = txid v = outorderid
	MERGE_LOCK     = "%s_merge_lock"   //每个地址合并限制5分钟

	BCH_RATE       = "bch_rate"
	LTC_RATE       = "ltc_rate"
	UCA_RATE       = "uca_rate"
	BCH_UTXO_LOCK  = "bch_utxo_lock_"
	LTC_UTXO_LOCK  = "ltc_utxo_lock_"
	UCA_UTXO_LOCK  = "uca_utxo_lock_"
	DCR_UTXO_LOCK  = "dcr_utxo_lock"
	Hc_UTXO_LOCK   = "hc_utxo_lock_"   //
	DOGE_UTXO_LOCK = "doge_utxo_lock_" //
	AVAX_UTXO_LOCK = "avax_utxo_lock_"
)

const (
	BTC_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
	USDT_UTXO_LOCK_SECOND_TIME = 2 * 60 //2分钟
	MERGE_LOCK_SECOND_TIME     = 2 * 60 //2分钟
	Hc_UTXO_LOCK_SECOND_TIME   = 3 * 60 //2分钟

	BCH_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
	LTC_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
	UCA_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
	DCR_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
	DOGE_UTXO_LOCK_SECOND_TIME = 2 * 60 //2分钟
	AVAX_UTXO_LOCK_SECOND_TIME = 2 * 60 //2分钟
	BIW_UTXO_LOCK_SECOND_TIME  = 2 * 60 //2分钟
)
