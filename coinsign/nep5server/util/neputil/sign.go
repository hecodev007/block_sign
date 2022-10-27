package neputil

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/O3Labs/neo-utils/neoutils"
	"github.com/group-coldwallet/nep5server/util"
	"github.com/zwjlink/neo-thinsdk-go/neo"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

var snowFlake *util.SnowFlake

func init() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	snowFlake = util.NewSnowFlake(uint32(r.Int31n(100)), false)
}

//nep5Script dc68c8a842d41e8d09e6ac439177f9cfa83a1d60
func Nep5Transfer(from, to, fromPrivKey, nep5Script string, amount int64) (raw string, txid string, err error) {
	if !strings.HasPrefix(from, "A") || !ValidateNEOAddress(from) {
		return "", "", errors.New("From address error")
	}
	if !strings.HasPrefix(to, "A") || !ValidateNEOAddress(to) {
		return "", "", errors.New("to address error")
	}
	if fromPrivKey == "" {
		return "", "", errors.New("miss fromPrivKey")
	}
	if nep5Script == "" {
		return "", "", errors.New("miss nep5 script")
	}
	if amount <= 0 {
		return "", "", errors.New("amount error")
	}
	params := &neo.CreateSignParams{
		Version: 0,
		PriKey:  fromPrivKey,
		From:    from,
		To:      to,
	}
	attrsData, ok := neo.GetPublicKeyHashFromAddress(params.From)
	if !ok {
		return "", "", errors.New("params.From scriptAddress error")
	}
	//引入一个第三方包做对比
	scriptAddr := neoutils.NEOAddressToScriptHashWithEndian(params.From, binary.LittleEndian)

	if hex.EncodeToString(attrsData) != scriptAddr {
		return "", "", errors.New("params.From scriptAddress error")
	}
	params.Attrs = []neo.Attribute{
		neo.Attribute{
			Usage: neo.Script,
			Data:  attrsData,
		},
	}
	var value = big.NewInt(amount)
	snowFlakeId, _ := snowFlake.Next()
	data, ok := neo.GetNep5Transfer(nep5Script, params.From, params.To, *value, *big.NewInt(int64(snowFlakeId)))
	if !ok {
		return "", "", errors.New("GetNep5Transfer data error")
	}
	params.Data = data
	txid, raw, err = neo.CreateNep5Tx(neo.InvocationTransaction, params)
	if err != nil {
		return "", "", err
	}
	return raw, txid, nil

}
