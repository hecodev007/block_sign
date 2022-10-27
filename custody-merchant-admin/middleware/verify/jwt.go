package verify

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/errcode"
	"custody-merchant-admin/util/library"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"strings"
)

func VerifyJWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			user := ctx.Get("user").(*jwt.Token)
			claims := user.Claims.(*domain.JwtCustomClaims)
			if strings.Contains(ctx.Request().URL.Path, "/api/open/v1") || strings.Contains(ctx.Request().URL.Path, "/api/rpc/v1") {
				return nil
			}
			token := ""
			cache.GetRedisClientConn().Get(global.GetCacheKey(global.UserTokenVerify, claims.Id), &token)
			if token != "" {
				t := token
				if t != user.Raw {
					ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
						"code": errcode.UnauthorizedAuthFail.StatusCode(),
						"msg":  "JWT已经失效",
					})
					return fmt.Errorf("JWT已经失效")
				}
				err := JwtSignVerify(ctx, token)
				if err != nil {
					// cdn错误
					ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
						"code": errcode.UnauthorizedAuthFail.StatusCode(),
						"msg":  err.Error(),
					})
					return err
				}
				return next(ctx)
			} else {
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "JWT权限验证没有通过",
				})
				return fmt.Errorf("JWT权限验证没有通过")
			}
		}
	}
}

func JwtSignVerify(ctx echo.Context, token string) error {
	var PassList = map[string]bool{
		"/admin/bill/export":         true,
		"/admin/orders/export":       true,
		"/admin/income/export":       true,
		"/admin/chain/bill/export":   true,
		"/admin/business/order/down": true,
	}
	if _, ok := PassList[ctx.Request().URL.Path]; ok {
		return nil
	}
	if Conf.JwtOut {
		req := ctx.Request()
		signStr := req.Header.Get(global.XCaSignStr)
		cts := ctx.Request().Header.Get(global.XCaTime)
		nonce := ctx.Request().Header.Get(global.XCaNonce)
		fmt.Printf("%s,%s,%s \n", signStr, cts, nonce)

		user := ctx.Get("user").(*jwt.Token)
		claims := user.Claims.(*domain.JwtCustomClaims)
		randomStr := []byte(claims.Nonce)
		str := []byte(fmt.Sprintf("Bearer %s%s%s", token, nonce, cts))
		signKey := library.AesEncryptECB(str, randomStr)
		fmt.Printf("randomStr,%s, \n", claims.Nonce)

		fmt.Println(base64.StdEncoding.EncodeToString(signKey))

		if base64.StdEncoding.EncodeToString(signKey) != signStr {
			return errors.New("异常签名")
		}
	}
	return nil
}
