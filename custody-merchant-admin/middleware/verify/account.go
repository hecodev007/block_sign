package verify

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/errcode"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func VerifyAccount() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			request := ctx.Request()
			// 获取请求的URI
			obj := request.URL.RequestURI()
			if obj == "/admin/base/reset/firstPassword" {
				return next(ctx)
			}
			user := ctx.Get("user").(*jwt.Token)
			claims := user.Claims.(*domain.JwtCustomClaims)
			token := false
			key := global.GetCacheKey(global.CheckAccount, claims.Id)
			cache.GetCacheStore().Get(key, &token)
			if token {
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "第一次登录,需要重置密码",
				})
				return fmt.Errorf("第一次登录,需要重置密码")

			}
			return next(ctx)
		}
	}
}
