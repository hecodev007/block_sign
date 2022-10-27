package crypto

import (
	_ed25519 "github.com/JFJun/helium-go/crypto/ed25519"
	nist_p256 "github.com/JFJun/helium-go/crypto/nist-p256"
)

/*
Helium 支持ed25519和NIST p-256两种非对称加密格式
*/

type Curves interface {
	GenerateKey() ([]byte, []byte)
	GetVersion() []byte
}

func NewCurve(version int) Curves {
	var c Curves
	if version == 0 {
		nc := &nist_p256.NISTP256Curve{Version: []byte{byte(version)}}
		c = nc
	} else if version == 1 {
		ec := &_ed25519.Ed25519Curve{Version: []byte{byte(version)}}
		c = ec
	}
	return c
}
