package util

import (
	"testing"
)

func TestAesEncode(t *testing.T) {
	aesKey := []byte("fqD4dzSn5s2yQJGQQ5yNekfG229vPqKz")
	t.Logf("随机生成密钥：%s", aesKey)
	crypter, _ := AesBase64Crypt([]byte("9462a47463c9de3dd0c144c29b8fe8744740f78c10fdf602ee8aa75be788b38d"), aesKey, true)
	t.Logf("加密结果：%s", string(crypter))
}

func TestAesDecode(t *testing.T) {
	ciphertext := "uiV1IcdYQ+W1GHhlpfqlRi5dB6tsWLpoOcb448tMbZdCXNBx0PIQDg2gry2VVlJXUoZHOQ=="
	key := "/+t4wzKOeS2qqcLTcbTQJi/ubtPQUt1b"
	result := "KyAevMVhccueJq9c4FiUUsXjQJudXWHHCLDeXaCMbx4JAnTY3VgK"

	text, _ := AesBase64Crypt([]byte(ciphertext), []byte(key), false)
	t.Logf("解密结果：%s", string(text))
	t.Logf("解密结果对比：%t", result == string(text))
}

func TestAesDecode2(t *testing.T) {
	ciphertext := "4vP39OothMFEEsexZUgtSpGp8Rphk5FBXoyJK/GRuHTFPDPMw3yOyYGF"
	key := "YHXLwu/MOVZ6OeSvMKnF7x6h8dyyuqOB"
	result := "0x2c8f4394D20C46C09FB493dFa2DC0B1b86116406"

	text, _ := AesBase64Crypt([]byte(ciphertext), []byte(key), false)
	t.Logf("解密结果：%s", string(text))
	t.Logf("解密结果对比：%t", result == string(text))
}

func TestAesEncry(t *testing.T) {
	aesKey := []byte("YHXLwu/MOVZ6OeSvMKnF7x6h8dyyuqOB")
	crypter, _ := AesBase64Crypt([]byte("349529c3cd288d272fd57a8dfaa922af1999b0d37c27e37e274155cc90f3bc23"), aesKey, true)
	t.Logf("加密结果：%s", string(crypter))
}
