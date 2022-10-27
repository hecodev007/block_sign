package auth

import (
	"custody-merchant-admin/internal/domain"
	"github.com/golang-jwt/jwt"
	"time"
)

const PrivateKey = "-----BEGIN RSA PRIVATE KEY-----\nMIICXgIBAAKBgQDXmqwFmi9TPVL1NgYZgGaHdPz7ZsskAOukaYg/hZkAn4rj8JJcaE/DzRy707bq6YBUpp+1ssiaadJRKR6+XMSzdyJz44Vl5rrArExrzGiNry+txRVo00U4xbIrmMX+aRKvPQ/xJ7xkxRhdxsY4F/wMcWDyJS+eyxJkHiBrHJ3ciQIDAQABAoGBAJV7eZUQx4sQ03mLkUMREQUNiXDMXj+CG96MBJj2CZSzCNrsqq1C7Tq19RwMt5+7cOw/8i9J22ejwtvehKA7NWxpsUBC+lDqXk3FCvtbL3d2fcARdh/1zWZN9WRvafkVPNPAeRC6ARp63DOe8FkT0C22DTOd0Xyvo0Zp7pF/GjXhAkEA/BkAFlV/4jHEglyXHNGrReMjClw2ClqKK5VXIk6UCJfVaGNDGbfw0ueYFnnOeIo8GPhgVjSC4wU2rX89pSFxTQJBANrxDqKc6wFw1jGpmxI25inxYTvA3SuSk36b4CSrRL7w3g9r+6QQfAlpBRZ9NBCL9WHeWHtgauxeDGJB2kmXui0CQQDhadlSHxFCSA3WIsRb2H609uwWD22ixGJXpilLW8eyB1GjDV6qWHbVno+3SSL9VV13Vl+NtVZzd+30JJoSVVzhAkB8sISxP8TnUSfrqLhUK0fx4zKJIVHUmum9VXDV8WR5ihwtlEYALhM2GMV5BV09fzgEwOiLe2Hps7ZBz1dOSkcRAkEA+D3kzvNpEYtqpjGHfUCxwmu/BwathDo09vj+gCcjhoJh/ADpa8+a0RQA6vcVMges0UcmIiIyQPNzCGlLBXtl9A==\n-----END RSA PRIVATE KEY-----"

func GetJWT(djwt *domain.JwtCustomClaims) (string, error) {
	djwt.StandardClaims = jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 8).Unix(),
	}
	// Set custom claims
	claims := djwt
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	return token.SignedString([]byte(PrivateKey))
}
