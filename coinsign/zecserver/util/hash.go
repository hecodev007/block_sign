package util

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
)

func Sha1HashString(data []byte) string {
	h := sha1.New()
	h.Write(data)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func Md5HashString(data []byte) string {
	signByte := []byte(data)
	hash := md5.New()
	hash.Write(signByte)
	return hex.EncodeToString(hash.Sum(nil))
}

//func Hash160String(data []byte) string {
//	return hex.EncodeToString(Hash160(data))
//}
//
//// Calculate the hash of hasher over buf.
//func calcHash(buf []byte, hasher hash.Hash) []byte {
//	hasher.Write(buf)
//	return hasher.Sum(nil)
//}
//
//// Hash160 calculates the hash ripemd160(sha256(b)).
//func Hash160(buf []byte) []byte {
//	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
//}
