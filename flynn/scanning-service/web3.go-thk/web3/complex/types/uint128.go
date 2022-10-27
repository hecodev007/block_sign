package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	lessThan = iota - 1
	equal
	greaterThan

	Len = 32
)

// Big endian uint128
type Uint128 struct {
	H, L uint64
}

func (u *Uint128) Compare(o *Uint128) int {
	if u.H < o.H {
		return lessThan
	} else if u.H > o.H {
		return greaterThan
	}

	if u.L < o.L {
		return lessThan
	} else if u.L > o.L {
		return greaterThan
	}

	return equal
}

func (u *Uint128) And(o *Uint128) {
	u.H &= o.H
	u.L &= o.L
}

func (u *Uint128) Or(o *Uint128) {
	u.H |= o.H
	u.L |= u.L
}

func (u *Uint128) Xor(o *Uint128) {
	u.H ^= o.H
	u.L ^= o.L
}

func (u *Uint128) Add(o *Uint128) {
	carry := u.L
	u.L += o.L
	u.H += o.H

	if u.L < carry {
		u.H += 1
	}
}

func NewFromString(s string) (u *Uint128, err error) {

	if len(s) > Len {
		return nil, fmt.Errorf("s:%s length greater than 32", s)
	}

	b, err := hex.DecodeString(fmt.Sprintf("%032s", s))
	if err != nil {
		return nil, err
	}
	rdr := bytes.NewReader(b)
	u = new(Uint128)
	err = binary.Read(rdr, binary.BigEndian, u)
	return
}

func (u *Uint128) HexString() string {
	if u.H == 0 {
		return fmt.Sprintf("%x", u.L)
	}
	return fmt.Sprintf("%x%016x", u.H, u.L)
}

func (u *Uint128) String() string {
	return fmt.Sprintf("0x%032x", u.HexString())
}