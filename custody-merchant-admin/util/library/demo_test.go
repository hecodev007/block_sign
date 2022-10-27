package library

import (
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"testing"
)

//测试加密函数
func TestEncryptPassword(t *testing.T) {
	fmt.Println(EncryptSha256Password("yxh402063397", ""))
}

func TestCacheMax(t *testing.T) {
	key := xkutils.NewUUId("test")
	fmt.Println(key)
}

func TestMQ(t *testing.T) {

}
