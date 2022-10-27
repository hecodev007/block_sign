package util

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"math/big"
	"reflect"

	"github.com/btcsuite/btcutil/base58"
)

const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

//B58cencode encodes a byte array into base58 with prefix
func B58cencode(payload []byte, prefix []byte) string {
	n := make([]byte, (len(prefix) + len(payload)))
	copy(n, append(prefix, payload...))
	b58c := XTZEncode(n)
	return b58c
}

// B58cdecode -
func B58cdecode(payload string, prefix []byte) []byte {
	b58c, _ := XTZDecode(payload)
	return b58c[len(prefix):]
}

// Encode -
func XTZEncode(dataBytes []byte) string {

	// Performing SHA256 twice
	sha256hash := sha256.New()
	sha256hash.Write(dataBytes)
	middleHash := sha256hash.Sum(nil)
	sha256hash = sha256.New()
	sha256hash.Write(middleHash)
	hash := sha256hash.Sum(nil)

	checksum := hash[:4]
	dataBytes = append(dataBytes, checksum...)

	// For all the "00" versions or any prepended zeros as base58 removes them
	zeroCount := 0
	for _, b := range dataBytes {
		if b == 0 {
			zeroCount++
		} else {
			break
		}
	}

	// Performing base58 encoding
	encoded := base58.Encode(dataBytes)

	for i := 0; i < zeroCount; i++ {
		encoded = "1" + encoded
	}

	return encoded
}

// Decode -
func XTZDecode(encoded string) ([]byte, error) {
	zeroCount := 0
	for i := 0; i < len(encoded); i++ {
		if encoded[i] == 49 {
			zeroCount++
		} else {
			break
		}
	}

	dataBytes, err := b58decode(encoded)
	if err != nil {
		return []byte{}, err
	}

	if len(dataBytes) <= 4 {
		return []byte{}, errors.New("invalid decode length")
	}
	data, checksum := dataBytes[:len(dataBytes)-4], dataBytes[len(dataBytes)-4:]

	for i := 0; i < zeroCount; i++ {
		data = append([]byte{0}, data...)
	}

	// Performing SHA256 twice to validate checksum
	sha256hash := sha256.New()
	sha256hash.Write(data)
	middleHash := sha256hash.Sum(nil)
	sha256hash = sha256.New()
	sha256hash.Write(middleHash)
	hash := sha256hash.Sum(nil)

	if !reflect.DeepEqual(checksum, hash[:4]) {
		return []byte{}, errors.New("data and checksum don't match")
	}

	return data, nil
}

func b58decode(data string) ([]byte, error) {
	decimalData := new(big.Int)
	alphabetBytes := []byte(alphabet)
	multiplier := big.NewInt(58)

	for _, value := range data {
		pos := bytes.IndexByte(alphabetBytes, byte(value))
		if pos == -1 {
			return nil, errors.New("character not found in alphabet")
		}
		decimalData.Mul(decimalData, multiplier)
		decimalData.Add(decimalData, big.NewInt(int64(pos)))
	}
	return decimalData.Bytes(), nil
}
