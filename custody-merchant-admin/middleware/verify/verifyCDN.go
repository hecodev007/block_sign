package verify

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/errcode"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"strings"
	"time"
)

func VerifyCDN() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var PassList = map[string]bool{
				"/admin/bill/export":         true,
				"/admin/orders/export":       true,
				"/admin/income/export":       true,
				"/admin/chain/bill/export":   true,
				"/admin/business/order/down": true,
				"/health":                    true,
			}

			values := 0
			if !Conf.CdnDisable {
				return next(ctx)
			}

			if strings.Contains(ctx.Request().URL.Path, "/api/open/v1") || strings.Contains(ctx.Request().URL.Path, "/api/rpc/v1") {
				return next(ctx)
			}

			if _, ok := PassList[ctx.Request().URL.Path]; ok {
				return next(ctx)
			}

			nonce := ctx.Request().Header.Get(global.XCaNonce)
			cts := ctx.Request().Header.Get(global.XCaTime)
			if cts == "" {
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "请求头参数缺失",
				})
				return errors.New("请求头参数缺失")
			}
			ts := xkutils.StrToInt64(cts)
			nowTs := time.Now().Unix()
			// 单位秒
			if (nowTs-ts) > 60 || (nowTs-ts) < -30 {
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "请求时间过期",
				})
				return errors.New("nonce错误")
			}
			if nonce == "" || len(nonce) > 128 {
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "nonce错误",
				})
				return errors.New("nonce错误")
			}
			key := global.GetCacheKey(global.CdnNonce, nonce)
			cache.GetRedisClientConn().Get(key, &values)
			if values != 0 {
				// cdn错误
				ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), map[string]interface{}{
					"code": errcode.UnauthorizedAuthFail.StatusCode(),
					"msg":  "cdn错误",
				})
				return fmt.Errorf("cdn错误")
			}

			// 60，不允许重复
			cache.GetRedisClientConn().Set(key, 1, time.Second*60)
			return next(ctx)
		}
	}
}
