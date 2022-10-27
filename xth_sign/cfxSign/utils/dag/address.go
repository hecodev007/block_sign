package dag

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"strings"
)
func HashToAddr(addrHash string)(string,error){
	addrHash = strings.TrimPrefix(addrHash,"0x")
	hash20,err := hex.DecodeString(addrHash)
	if err != nil{
		return "",err
	}
	prehash20,err :=hex.DecodeString("030618"+addrHash)
	if err != nil {
		return "",err
	}
	sum := Sha2Sum(prehash20[:])
	hash20 = append(hash20,sum[0:4]...)
	addr :=base32.NewEncoding("abcdefghjkmnprstuvwxyz0123456789").EncodeToString(hash20)
	return "cfx:"+addr,nil
}

func Sha2Sum(b []byte) (out [32]byte) {
	ShaHash(b, out[:])
	return
}
func ShaHash(b []byte, out []byte) {
	s := sha256.New()
	s.Write(b[:])
	tmp := s.Sum(nil)
	s.Reset()
	s.Write(tmp)
	copy(out[:], s.Sum(nil))
}