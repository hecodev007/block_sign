package launcher

import "github.com/group-coldwallet/cocos/service/cocosImpl"

func UnlockWallet() {
	cocosImpl.Unlock()
}
