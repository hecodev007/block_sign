package neputil

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestNep5Transfer(t *testing.T) {

	//EcDhVcP84aGXZqphYPRqgonmtUyvvPowF6ZKDuDAqRPD
	//t.Log(
	//	Nep5Transfer(
	//		"AThW78vaTVG2gaQEQ6irJbwGJU22HhQdZw",
	//		"AaAPzrgH9RZMmrT1KnxJyHaYFLc1cNztd1",
	//		"",
	//		"ab38352559b8b203bde5fddfa0b07d8b2525e132",
	//		decimal.NewFromFloat(1173474.506).Shift(8).IntPart(),
	//	),
	//)

	t.Log(
		Nep5Transfer(
			"AdXXbNqTW5GDhezvnNR1L7hXnLxbs3mr3c",
			"AdGbMkB4dg8GbGdMDopoWZUXNvbxpa4zME",
			"",
			"3e09e602eeeb401a2fec8e8ea137d59aae54a139",
			decimal.NewFromFloat(19032.36).Shift(8).IntPart(),
		),
	)

}
