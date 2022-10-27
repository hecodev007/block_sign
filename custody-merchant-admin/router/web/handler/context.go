package handler

import (
	"bytes"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/middleware/cache"
	ot "custody-merchant-admin/middleware/opentracing"
	"custody-merchant-admin/middleware/session"
	"custody-merchant-admin/util/xkutils"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"io/ioutil"
	"sync"
)

var (
	ctxPool = sync.Pool{
		New: func() interface{} {
			return &Context{}
		},
	}
	DefaultEchoBinder *echo.DefaultBinder
)

func NewContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := ctxPool.Get().(*Context)
			defer func() {
				ctx.reset()
				ctxPool.Put(ctx)
			}()

			ctx.Context = c
			return next(ctx)
		}
	}
}

type Context struct {
	echo.Context
}

func (ctx *Context) reset() {
	ctx.Context = nil
}
func (ctx *Context) OpenTracingSpan() opentracing.Span {
	return ot.Default(ctx)
}

func (ctx *Context) Session() session.Session {
	return session.Default(ctx)
}

func (ctx *Context) GetTokenUser() *domain.JwtCustomClaims {
	user := ctx.Get("user").(*jwt.Token)
	return user.Claims.(*domain.JwtCustomClaims)
}

func (ctx *Context) TokenUserOut() *domain.JwtCustomClaims {
	user := ctx.Get("user").(*jwt.Token)
	ctx.Set("user", user)
	claim := user.Claims.(*domain.JwtCustomClaims)
	cache.GetRedisClientConn().Del(global.GetCacheKey(global.UserTokenVerify, claim.Id))
	return claim
}

func (ctx *Context) DefaultBinder(v interface{}) (err error) {
	if DefaultEchoBinder == nil {
		DefaultEchoBinder = &echo.DefaultBinder{}
	}
	err = DefaultEchoBinder.BindBody(ctx, v)
	if err != nil {
		return err
	}
	return err
}

func (ctx *Context) DataBinder(v interface{}) (err error) {
	data, _ := ioutil.ReadAll(ctx.Request().Body)
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(data))
	//var req = map[string]interface{}{}
	//json.Unmarshal(data, &req)
	ctx.Binder(v)
	return nil
}

func (ctx *Context) Binder(v interface{}) (err error) {
	if DefaultEchoBinder == nil {
		DefaultEchoBinder = &echo.DefaultBinder{}
	}
	err = DefaultEchoBinder.Bind(v, ctx)
	if err != nil {
		return err
	}
	return err
}

func (ctx *Context) DefaultQueryParams(v interface{}) (err error) {
	if DefaultEchoBinder == nil {
		DefaultEchoBinder = new(echo.DefaultBinder)
	}
	err = DefaultEchoBinder.BindQueryParams(ctx, v)
	if err != nil {
		return err
	}
	return err
}

func (ctx *Context) SwitchType(key, t string) interface{} {
	switch t {
	case "int":
		return xkutils.StrToInt(ctx.QueryParam(key))
	case "int64":
		return xkutils.StrToInt64(ctx.QueryParam(key))
	case "float64":
		return xkutils.StrToFloat64(ctx.QueryParam(key))
	}
	return ctx.QueryParam(key)
}

func (ctx *Context) OffsetPage() (int, int) {
	limit := xkutils.StrToInt(ctx.QueryParam("limit"))
	offset := xkutils.StrToInt(ctx.QueryParam("offset"))
	if limit <= 0 {
		limit = 10
	}
	if offset <= 0 {
		offset = 0
	}

	return limit * offset, limit
}
