package stx

import (
	"errors"
	"crypto/sha256"
	"github.com/ilius/crock32"
)
const AlphabetUpper= "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
func FromString(addr string)([]byte,error){
	return crock32.Decode(addr)
}
//version,hash转address
func C32_check_encode(version uint8,data []byte) (string,error){
	if version>=32{
		return "", errors.New("无效的version")
	}
	var check_data []byte
	check_data = append(check_data,version)
	check_data = append(check_data,data...)
	sum1 :=sha256.Sum256(check_data)
	sum2:=sha256.Sum256(sum1[:])
	checksum :=sum2[0:4]
	encoding_data:=append(data,checksum...)
	c32_string :=crock32.Encode(encoding_data)
	return "S"+string(AlphabetUpper[version])+c32_string,nil
}

func C32_check_decode(addr string) (version uint8,hash []byte,err error){
	hash,err =crock32.Decode(addr[2:])
	if err != nil {
		return
	}
	hash = hash[0:len(hash)-4]
	v,err := crock32.Decode(addr[1:2])
	if err != nil {
		return
	}
	version=v[0]
	return
}