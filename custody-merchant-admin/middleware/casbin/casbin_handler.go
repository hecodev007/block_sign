package middleware

import (
	conf "custody-merchant-admin/config"
	"custody-merchant-admin/db"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/module/errcode"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func CasbinHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if conf.Conf.Mod == "pro" {
				user := ctx.Get("user").(*jwt.Token)
				claims := user.Claims.(*domain.JwtCustomClaims)
				request := ctx.Request()
				if claims.Admin || claims.Role == 2 {
					return next(ctx)
				}
				// 获取请求的URI
				obj := request.URL.RequestURI()
				// 获取请求方法
				act := request.Method
				// 获取用户
				sub := fmt.Sprintf("%d", claims.Id)
				e := db.CasbinDB()
				// 判断策略中是否存在
				success, _ := e.Enforce(sub, obj, act)
				if success {
					//log.Println("恭喜您,权限验证通过")
					return next(ctx)
				} else {
					//ctx.Error(errcode.UnauthorizedAuthFail)
					data := map[string]interface{}{}
					data["msg"] = "很遗憾,权限验证没有通过"
					ctx.JSON(errcode.UnauthorizedAuthFail.StatusCode(), data)
					return fmt.Errorf("很遗憾,权限验证没有通过")
				}
			}
			return next(ctx)
		}
	}
}
