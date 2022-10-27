package dogeutil

import (
	"testing"
)

//D8QP6xz5UUzmkckbH4LHenJNVBSQTcCQ93
//QUa6WqF3tFDgMdUF9emSghdPY1rzexsXjJawbNE9VrAkrWJJiMw3

//DEcvyXat5AVMPpDQgRoeQvym6PJGwsXjQd
//QR1JHJVRYgYKEhvCFa48pHyTk36kcjSQxVQfx9ZyXS4a9XhFjA2H

//DD1PG6X57ovwUNAKt3c8DUavv1szWwKSwR
//QS32wB2BrybqDFSut2VSfHczNaKXKSphxQJXzRCduvxMtJmRUujX

func TestAddress(t *testing.T) {
	for i := 0; i < 3; i++ {
		//wif, _ := CreatePrivateKey()
		////wif, _ := ImportWIF("your compressed privateKey Wif")
		////wif, _ := ImportWIF("L3sy1eqQJKkz2kjShUukLpS1zGd2ZoGrBz9uupiP7sUYWhu44KVg")
		//
		//address, _ := GetAddress(wif)
		//t.Log("Common Address:", address.EncodeAddress())
		//t.Log("PrivateKeyWifCompressed:", wif.String())

		t.Log(CreateAddr())
	}

}
