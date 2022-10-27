package neo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/o3labs/neo-utils/neoutils/btckey"
)

func BytesToNeoAddr(bts string) string {
	bt, _ := hex.DecodeString(bts)
	return btckey.B58checkencodeNEO(0x17, bt)
}
func bytesToInt(he string) (int64, error) {
	if len(he)%2 == 1 {
		he = "0" + he
	}
	if len(he) > 16 {
		for i := 16; i < len(he); i++ {
			if he[i] != '0' {
				return 0, fmt.Errorf("data too long")
			}
		}
		he = he[:16]
	}

	for len(he) < 16 {
		he = he + "0"
	}
	b, err := hex.DecodeString(he)
	if err != nil {
		return 0, err
	}
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint64
	err = binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return int64(tmp), err
}
