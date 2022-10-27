package redis

import (
	"fmt"
	"strings"
)

const (
	CacheKeyForceRePush      = "force_repush_"
	CacheKeyWaitingOrderList = "bc_wait_trx_order_list"
	CacheKeyTxProcessLock    = "bc_tx_process_lock"
	WatchSameWayBackCacheKey = "swb"
	PrepareNotifyCacheKey    = "prepare_notify_reload_set"
)

func GetPrepareNotifyCacheKey() string {
	return PrepareNotifyCacheKey
}

func GetCollectingCacheKey(outerOrderNo string) string {
	return fmt.Sprintf("%s_%s", collectingCacheKey, outerOrderNo)
}

func GetRePushCacheKey(outerOrderNo string) string {
	return fmt.Sprintf("%s_%s", repushCacheKey, outerOrderNo)
}

func GetTxProcessLockCacheKey(outerOrderNo string, mch string) string {
	return fmt.Sprintf("%s_%d_%s", CacheKeyTxProcessLock, mch, outerOrderNo)
}

func GetWatchSameWayBackCacheKey(chain, coinCode, addr1, addr2 string) string {
	coin := chain
	if coinCode != "" {
		coin = coinCode
	}
	return GetWatchSameWayBackCacheKeyWithCoinType(coin, addr1, addr2)
}

func GetWatchSameWayBackCacheKeyWithCoinType(coinType, addr1, addr2 string) string {
	return fmt.Sprintf("%s_%s_%s_%s", WatchSameWayBackCacheKey, strings.ToLower(coinType), strings.ToLower(addr1), strings.ToLower(addr2))
}
