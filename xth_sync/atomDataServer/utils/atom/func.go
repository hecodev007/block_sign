package atom

import (
	"errors"
	"github.com/shopspring/decimal"
	"strings"
)

func AtomToInt(value string) (int64, error) {
	if value == "" || value == "0" || value=="0uatom,0uatom"{
		return 0, nil
	}
	if strings.HasSuffix(value,"atoms"){
		return 0,nil
	}
	//10uatom = 0.000010
	if len(value) >= 5 && (value[len(value)-5:len(value)] == "uatom") {
		value, err := decimal.NewFromString(value[0 : len(value)-5])
		if err != nil {
			return 0, err
		}
		return value.IntPart(), nil
		//10atom = 10.000000
	} else if len(value) >= 4 && value[len(value)-4:len(value)] == "atom" {
		value, err := decimal.NewFromString(value[0 : len(value)-4])
		if err != nil {
			return 0, err
		}
		return value.Shift(6).IntPart(), nil
		//10utaom = 10.000000
		//} else if len(value) >= 5 && value[len(value)-5:len(value)] == "utaom" {
		//	value, err := decimal.NewFromString(value[0 : len(value)-5])
		//	if err != nil {
		//		return 0, err
		//	}
		//	return value.Shift(6).IntPart(), nil
	} else {
		return 0,nil
		return 0, errors.New("需要该逻辑处理" + " " + value)
	}
}
