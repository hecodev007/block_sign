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
