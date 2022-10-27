package models

import (
	util "lunasign/utils/luna"
	"lunasign/utils/keystore"
	"testing"
)

func Test_atom(t *testing.T) {
	address, private, err := util.GenAccount()
	if err != nil {
		return
	}
	var cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD []*keystore.CsvKey
	aesKey := keystore.RandBase64Key()
	aesPrivKey, err := keystore.AesBase64CryptCfb([]byte(private), aesKey, true)
	if err != nil {
		return
	}
	cvsKeysA = append(cvsKeysA, &keystore.CsvKey{Address: address, Key: string(aesPrivKey)})
	cvsKeysB = append(cvsKeysB, &keystore.CsvKey{Address: address, Key: string(aesKey)})
	cvsKeysC = append(cvsKeysC, &keystore.CsvKey{Address: address, Key: string(keystore.Base64Encode([]byte(private)))}) //string(keystore.Base64Encode([]byte(private)))})
	cvsKeysD = append(cvsKeysD, &keystore.CsvKey{Address: address, Key: ""})

	if err := keystore.GenerateCvsABC(cvsKeysA, cvsKeysB, cvsKeysC, cvsKeysD, "main", "123"); err != nil {
		panic("")
	}

}
