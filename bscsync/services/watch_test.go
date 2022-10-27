package services

import (
	"dataserver/conf"
	"dataserver/utils"
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	a, _ := utils.AesBase64Str("6knmqfRSBnm£nKNoKqgoC284lriwMppBmnYedYdOK9LFRdab8+F9Fo7uT1PsVaW/XnGhtL5xTqCA=", conf.DefaulAesKey, false)
	fmt.Println(a)

	b, _ := utils.AesBase64Str("70bxsqQQRCW9Jd4eqUEJh6+rxQyAm/u6l7UHpyET8UsjYHSthv5OeRx3jFjl7iccw1ucPR/26dIuNf0=", conf.DefaulAesKey, false)
	fmt.Println(b)

	c, _ := utils.AesBase64Str("70XjuJEOTHmoLdAB", conf.DefaulAesKey, false)
	fmt.Println(c)

	d, _ := utils.AesBase64Str("32in/6k/X2K8fNwO+UIG3Q==", conf.DefaulAesKey, false)
	fmt.Println(d)
}
