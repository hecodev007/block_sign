// Package pseudohsm provides a pseudo HSM for development environments.
package pseudohsm

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"btmSign/bytom/crypto/ed25519/chainkd"
)

const logModule = "pseudohsm"

// KeyImage is the struct for hold export key data
type KeyImage struct {
	XKeys []*encryptedKeyJSON `json:"xkeys"`
}

// Backup export all the HSM keys into array
func (h *HSM) Backup() (*KeyImage, error) {
	image := &KeyImage{}
	xpubs := h.cache.keys()
	for _, xpub := range xpubs {
		data, err := ioutil.ReadFile(xpub.File)
		if err != nil {
			return nil, err
		}

		xKey := &encryptedKeyJSON{}
		if err := json.Unmarshal(data, xKey); err != nil {
			return nil, err
		}

		image.XKeys = append(image.XKeys, xKey)
	}
	return image, nil
}

// Restore import the keyImages into HSM
func (h *HSM) Restore(image *KeyImage) error {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	for _, xKey := range image.XKeys {
		data, err := hex.DecodeString(xKey.XPub)
		if err != nil {
			return ErrXPubFormat
		}

		var xPub chainkd.XPub
		copy(xPub[:], data)
		if h.cache.hasKey(xPub) {
			log.WithFields(log.Fields{
				"module": logModule,
				"alias":  xKey.Alias,
				"id":     xKey.ID,
				"xPub":   xKey.XPub,
			}).Warning("skip restore key due to already existed")
			continue
		}

		if ok := h.cache.hasAlias(xKey.Alias); ok {
			return ErrDuplicateKeyAlias
		}

		rawKey, err := json.Marshal(xKey)
		if err != nil {
			return err
		}

		_, fileName := filepath.Split(xKey.ID)
		file := h.keyStore.JoinPath(keyFileName(fileName))
		if err := writeKeyFile(file, rawKey); err != nil {
			return err
		}

		h.cache.reload()
	}
	return nil
}
