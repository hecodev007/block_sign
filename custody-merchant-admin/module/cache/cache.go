package cache

import (
	"time"

	"github.com/labstack/echo/v4"

	. "custody-merchant-admin/config"
	ec "custody-merchant-admin/middleware/cache"
)

const (
	DefaultExpiration = 3600
	DEFAULT           = time.Duration(0)
	FOREVER           = time.Duration(-1)
	DefaultKey        = "cache"
)

func Cache() echo.MiddlewareFunc {
	var store ec.CacheStore
	switch Conf.CacheStore {
	case MEMCACHED:
		store = ec.NewMemcachedStore([]string{Conf.Memcached.Server}, time.Hour)
	case REDIS:
		store = ec.NewRedisCache(Conf.Redis.AloneAddress, Conf.Redis.AlonePwd, DefaultExpiration)
	default:
		store = ec.NewInMemoryStore(time.Hour)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(DefaultKey, store)
			return next(c)
		}
	}
}

// Default shortcut to get Cache
func Default(c echo.Context) ec.CacheStore {
	// return c.MustGet(DefaultKey).(ec.CacheStore)
	return c.Get(DefaultKey).(ec.CacheStore)
}
