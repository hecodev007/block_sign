package domain

import (
	"github.com/golang-jwt/jwt"
)

// JwtCustomClaims are custom claims extending default ones.
// See https://github.com/golang-jwt/jwt for more examples
type JwtCustomClaims struct {
	Admin      bool   `json:"admin"`
	Id         int64  `json:"id"`
	MerchantId int64  `json:"merchant_id"`
	Role       int    `json:"role"`
	Name       string `json:"name"`
	Account    string `json:"account"`
	Nonce      string `json:"nonce"`
	jwt.StandardClaims
}
