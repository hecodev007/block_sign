package utils

import "encoding/base64"

// EncodeToBase64 encodes a byte array to base64 raw url encoding
func EncodeToBase64(toEncode []byte) string {
	return base64.RawURLEncoding.EncodeToString(toEncode)
}

// DecodeString decodes from base64 raw url encoding to byte array
func DecodeString(toDecode string) ([]byte, error) {
	if toDecode == "" {
		return []byte{}, nil
	}
	return base64.RawURLEncoding.DecodeString(toDecode)
}

func StringToBuffer(data string) []byte {
	return []byte(data)
}
