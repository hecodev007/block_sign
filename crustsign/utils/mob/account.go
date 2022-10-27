package mob

import (
	"crypto/rand"
	"github.com/bwesterb/go-ristretto"
	"io"
)

func GenAccount()(addr string,private string,err error){
	var r [32]byte
	_, err = io.ReadFull(rand.Reader,r[:])
	if err != nil {
		return "", "", err
	}
	scalar := new(ristretto.Scalar)
	scalar.SetBytes(&r)
	return scalar.String(),"",nil


}