package mw

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"golang.org/x/crypto/curve25519"
)

func GenAccount() (address string, private string, err error) {
	var priv, pubKey [32]byte
	_, err = rand.Reader.Read(priv[:])
	if err != nil {
		return "", "", err
	}

	curve25519.ScalarBaseMult(&pubKey, &priv)
	address, err = PubkeyToAddr(hex.EncodeToString(pubKey[:]))
	return address, hex.EncodeToString(priv[:]), err
}

//
//func sign(h, x, s [32]byte) []byte {
//	h1 := make([]byte, 32)
//	x1 := make([]byte, 32)
//	tmp1 := make([]byte, 64)
//	tmp2 := make([]byte, 64)
//	var w, i int
//	copy(h1, h[:])
//	copy(x1, x[:])
//	tmp3 := make([]byte, 32)
//	divmod(tmp3, h1, 32, ORDER, 32)
//	divmod(tmp3, x1, 32, ORDER, 32)
//
//	v := make([]byte, 32)
//	mula_small2(v, x1, 0, h1, 32, -1)
//	mula_small(v, v, 0, ORDER, 32, 1)
//
//	// tmp1 = (x-h)*s mod q
//	mula32(tmp1, v, s[:], 32, 1)
//	divmod(tmp2, tmp1, 64, ORDER, 32)
//	w = 0
//	i = 0
//	for ; i < 32; i++ {
//		v[i] = tmp1[i]
//		w |= int(v[i])
//	}
//	if w != 0 {
//		return v
//	}
//	return nil
//}
//func mula32(p, x []byte, y []byte, t, z int) int {
//	t = t | 0
//	z = z | 0
//
//	var n = 31
//	var w = 0
//	var i = 0
//	for ; i < t; i++ {
//		zy := z * (int(y[i]) & 0xFF)
//		w += mula_small2(p, p, i, x, n, zy) + (int(p[i+n]) & 0xFF) + zy*(int(x[n])&0xFF)
//		p[i+n] = uint8(w & 0xFF)
//		w >>= 8
//	}
//	p[i+n] = uint8((w + int(p[i+n])&0xFF) & 0xFF)
//	return w >> 8
//}

//var ORDER = []int{
//	237, 211, 245, 92,
//	26, 99, 18, 88,
//	214, 156, 247, 162,
//	222, 249, 222, 20,
//	0, 0, 0, 0,
//	0, 0, 0, 0,
//	0, 0, 0, 0,
//	0, 0, 0, 16,
//}

//func divmod(q, r []byte, n int, d []int, t int) {
//	n = n | 0
//	t = t | 0
//
//	rn := 0
//	dt := (d[t-1] & 0xFF) << 8
//	if t > 1 {
//		dt |= (d[t-2] & 0xFF)
//	}
//	for n >= t {
//		n -= 1
//		z := (rn << 16) | ((int(r[n]) & 0xFF) << 8)
//		if n > 0 {
//			z |= (int(r[n-1]) & 0xFF)
//		}
//
//		i := n - t + 1
//		z /= dt
//		rn += mula_small(r, r, i, d, t, -z)
//		q[i] = uint8((z + rn) & 0xFF)
//		/* rn is 0 or -1 (underflow) */
//		mula_small(r, r, i, d, t, -rn)
//		rn = int(r[n]) & 0xFF
//		r[n] = 0
//	}
//
//	r[t-1] = uint8(rn & 0xFF)
//}
//func mula_small(p, q []byte, m int, x []int, n, z int) int {
//	m = m | 0
//	n = n | 0
//	z = z | 0
//
//	var v = 0
//	for i := 0; i < n; i++ {
//		v += (int(q[i+m]) & 0xFF) + z*(int(x[i])&0xFF)
//		p[i+m] = uint8(v & 0xFF)
//		v >>= 8
//	}
//
//	return v
//}
//func mula_small2(p, q []byte, m int, x []byte, n, z int) int {
//	m = m | 0
//	n = n | 0
//	z = z | 0
//
//	var v = 0
//	for i := 0; i < n; i++ {
//		v += (int(q[i+m]) & 0xFF) + z*(int(x[i])&0xFF)
//		p[i+m] = uint8(v & 0xFF)
//		v >>= 8
//	}
//
//	return v
//}
func PrivateToAddr(pri string) (string, error) {
	private, err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	var pubKey, priv [32]byte
	copy(priv[:], private)
	curve25519.ScalarBaseMult(&pubKey, &priv)

	return PubkeyToAddr(hex.EncodeToString(pubKey[:]))
}
func PrivateToPub(pri string) (string, error) {
	private, err := hex.DecodeString(pri)
	if err != nil {
		return "", err
	}
	var pubKey, priv [32]byte
	copy(priv[:], private)
	curve25519.ScalarBaseMult(&pubKey, &priv)
	return hex.EncodeToString(pubKey[:]), nil
}
func PubkeyToAddr(pub string) (string, error) {
	pubhash := Sha256(pub)
	bi := big.NewInt(0)
	bi.SetBytes(Converse(pubhash[0:8]))
	addr := newAddress()
	ret := addr.from_acc(bi.String())
	if !ret {
		return "", errors.New("invalid pubkey")
	}
	return addr.toString(), nil
}
func AddrToAccoutid(addr string) (string, error) {
	address := newAddress()
	if !address.fromRS(addr) {
		return "", errors.New("invalid address")
	}
	return address.account_id(), nil
}
func getPrivate(Phrase string) string {
	hash := Sha256(hex.EncodeToString([]byte(Phrase)))
	sthash := byteArrayToShortArray(hash)
	sthash = curve25519_clamp(sthash)

	return shortArrayToHexString(sthash)
}

func shortArrayToHexString(shotArray []uint16) string {
	byteArray := shortArrayToByteArray(shotArray)
	return hex.EncodeToString(byteArray)
}
func byteArrayToShortArray(byteArray []byte) []uint16 {
	shortArray := []uint16{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if len(byteArray) < 16 {
		return shortArray
	}
	for i := 0; i < 16; i++ {
		shortArray[i] = uint16(byteArray[i*2]) | uint16(byteArray[i*2+1])<<8
	}
	return shortArray
}
func shortArrayToByteArray(shortArray []uint16) []byte {
	byteArray := make([]byte, 32)
	if len(shortArray) < 8 {
		return byteArray
	}
	for i := 0; i < 16; i++ {
		byteArray[2*i] = uint8(shortArray[i] & 0xff)
		byteArray[2*i+1] = uint8(shortArray[i] >> 8)
	}
	return byteArray
}
func curve25519_clamp(curve []uint16) []uint16 {
	curve[0] &= 0xFFF8
	curve[15] &= 0x7FFF
	curve[15] |= 0x4000
	return curve
}

func Sha256(hexstring string) []byte {
	data, _ := hex.DecodeString(hexstring)
	ret := sha256.Sum256(data)
	return ret[:]
}
func Converse(d []byte) []byte {
	for i := 0; i < len(d)/2; i++ {
		d[i], d[len(d)-1-i] = d[len(d)-1-i], d[i]
	}
	return d
}

type mwaddress struct {
	codeword []int
	syndrome []int
	gexp     []int
	glog     []int
	cwmap    []int
	guess    []string
	alphabet string
}

func newAddress() *mwaddress {
	mw := new(mwaddress)
	mw.codeword = []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	mw.syndrome = []int{0, 0, 0, 0, 0}
	mw.gexp = []int{1, 2, 4, 8, 16, 5, 10, 20, 13, 26, 17, 7, 14, 28, 29, 31, 27, 19, 3, 6, 12, 24, 21, 15, 30, 25, 23, 11, 22, 9, 18, 1}
	mw.glog = []int{0, 0, 1, 18, 2, 5, 19, 11, 3, 29, 6, 27, 20, 8, 12, 23, 4, 10, 30, 17, 7, 22, 28, 26, 21, 25, 9, 16, 13, 14, 24, 15}
	mw.cwmap = []int{3, 2, 1, 0, 7, 6, 5, 4, 13, 14, 15, 16, 12, 8, 9, 10, 11}
	mw.alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	return mw
}
func (mw *mwaddress) set_codeword(cw []int) {
	len := 17
	skip := -1
	i := 0
	j := 0
	for ; i < len; i++ {
		if i != skip {
			mw.codeword[mw.cwmap[j]] = cw[i]
			j++
		}
	}
}

//aaaa1111bbbb2222
func (mw *mwaddress) set(adr string) *mwaddress {
	adr = adr[0:16]
	var clean [17]int
	for i := 0; i < 16; i++ {
		pos := mw.indexOf(adr[i])
		clean[i] = pos
	}
	for i := 16; i >= 0; i-- {
		for j := 0; j < 32; j++ {
			clean[i] = j

			mw.set_codeword(clean[:])

			if mw.ok() {
				mw.add_guess()
			}
		}

		if i > 0 {
			t := clean[i-1]
			clean[i-1] = clean[i]
			clean[i] = t
		}
	}
	mw.reset()
	return mw
}
func (mw *mwaddress) reset() {
	for i := 0; i < 17; i++ {
		mw.codeword[i] = 1
	}
}
func (mw *mwaddress) gmult(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}

	var idx = (mw.glog[a] + mw.glog[b]) % 31
	return mw.gexp[idx]
}
func (mw *mwaddress) ok() bool {
	var sum = 0

	for i := 1; i < 5; i++ {
		j := 0
		t := 0
		for ; j < 31; j++ {
			if j > 12 && j < 27 {
				continue
			}
			pos := j
			if j > 26 {
				pos -= 14
			}

			t ^= mw.gmult(mw.codeword[pos], mw.gexp[(i*j)%31])
		}

		sum |= t
		mw.syndrome[i] = t
	}

	return sum == 0
}
func (mw *mwaddress) add_guess() {
	s := mw.toString()
	length := len(mw.guess)

	if length > 2 {
		return
	}

	for i := 0; i < length; i++ {
		if mw.guess[i] == s {
			return
		}
	}

	mw.guess[length] = s
}
func (mw *mwaddress) toString() string {
	out := "CDW-"

	for i := 0; i < 17; i++ {
		out += mw.alphabet[mw.codeword[mw.cwmap[i]] : mw.codeword[mw.cwmap[i]]+1]

		if (i&3) == 3 && i < 13 {
			out += "-"
		}
	}

	return out
}

func (mw *mwaddress) indexOf(b uint8) int {
	for i := 0; i < len(mw.alphabet); i++ {
		if mw.alphabet[i] == b {
			return i
		}
	}
	return -1
}
func (mw *mwaddress) from_acc(acc string) bool {
	var inp [20]int
	var out [20]int
	var pos = 0
	var length = len(acc)

	if length == 20 && acc[0] != '1' {
		return false
	}

	for i := 0; i < length; i++ {
		inp[i] = int(acc[i] - '0')
	}

	for {
		divide := 0
		newlen := 0

		for i := 0; i < length; i++ {
			divide = divide*10 + inp[i]

			if divide >= 32 {
				inp[newlen] = divide >> 5
				newlen++
				divide &= 31
			} else if newlen > 0 {
				inp[newlen] = 0
				newlen++
			}
		}

		length = newlen
		out[pos] = divide
		pos++
		if newlen == 0 {
			break
		}
	}

	for i := 0; i < 13; i++ {
		pos--
		if pos >= 0 {
			mw.codeword[i] = out[i]
		} else {
			mw.codeword[i] = 0
		}
	}

	mw.encode()

	return true
}
func (mw *mwaddress) encode() {
	p := []int{0, 0, 0, 0}

	for i := 12; i >= 0; i-- {
		var fb = mw.codeword[i] ^ p[3]

		p[3] = p[2] ^ mw.gmult(30, fb)
		p[2] = p[1] ^ mw.gmult(6, fb)
		p[1] = p[0] ^ mw.gmult(9, fb)
		p[0] = mw.gmult(17, fb)
	}

	mw.codeword[13] = p[0]
	mw.codeword[14] = p[1]
	mw.codeword[15] = p[2]
	mw.codeword[16] = p[3]
}

func (mw *mwaddress) account_id() string {
	out := make([]byte, 0)
	inp := make([]int, 13, 13)
	length := 13
	for i := 0; i < 13; i++ {
		inp[i] = mw.codeword[12-i]
	}
	for {
		divide := 0
		newlen := 0
		for i := 0; i < length; i++ {

			divide = divide*32 + inp[i]
			if divide >= 10 {
				inp[newlen] = divide / 10
				newlen++
				divide %= 10
			} else if newlen > 0 {
				inp[newlen] = 0
				newlen++
			}
		}

		length = newlen
		out = append(out, uint8(divide)+'0')
		if newlen <= 0 {
			break
		}
	}
	for i := 0; i < len(out)/2; i++ {
		out[i], out[len(out)-1-i] = out[len(out)-1-i], out[i]
	}
	return string(out)
}
func (mw *mwaddress) fromRS(addr string) bool {
	if !strings.HasPrefix(addr, "CDW-") {
		return false
	}
	clean := make([]int, 0)
	for i := 4; i < len(addr); i++ {
		pos := mw.indexOf(addr[i])

		if pos >= 0 {
			clean = append(clean, int(pos))
			if len(clean) > 18 {
				return false
			}
		}
	}
	mw.set_codeword(clean)
	if mw.ok() {
		return true
	}
	mw.reset()
	return true
}
