package orm

import (
	"custody-merchant-admin/model"
	"reflect"
	"time"

	"gorm.io/gorm"

	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/conv"
	"custody-merchant-admin/util/crypt"
	"custody-merchant-admin/util/sql"
)

const (
	CacheExpireDefault = time.Minute
	CacheKeyFormat     = "SQL:%s SQLVars:%v"
)

type CacheDB struct {
	*gorm.DB
	store  cache.CacheStore
	Expire time.Duration
}

type CacheConf struct {
	Expire time.Duration
}

func Cache(db *gorm.DB) *CacheDB {
	return NewCacheDB(db, model.CacheStore(), CacheConf{
		Expire: time.Second * 10,
	})
}

func NewCacheDB(db *gorm.DB, store cache.CacheStore, conf CacheConf) *CacheDB {
	switch conf.Expire {
	case time.Duration(0):
		conf.Expire = CacheExpireDefault
	}

	newDB := CacheDB{
		DB:     db,
		store:  store,
		Expire: conf.Expire,
	}
	return &newDB
}

func (c *CacheDB) First(out interface{}, where ...interface{}) *CacheDB {
	sql := gorm.Statement{}
	key := ""
	key = cacheKey(sql)
	err := c.store.Get(key, out)
	if cache.IsNotHave(err) {
		log.Debugf("find no cache data")
		db := c.DB.First(out, where)
		if err := db.Error; err == nil {
			c.store.Set(key, out, c.Expire)
			return c
		}
	}
	c.DB = c.DB.First(&sql, out, where)
	return c
}

func (c *CacheDB) Last(out interface{}, where ...interface{}) *CacheDB {

	sql := gorm.Statement{}
	key := ""
	key = cacheKey(sql)
	err := c.store.Get(key, out)
	if cache.IsNotHave(err) {
		log.Debugf("find no cache data")
		db := c.DB.Last(out, where)
		if err := db.Error; err == nil {
			c.store.Set(key, out, c.Expire)
			return c
		}
	}
	c.DB = c.DB.Last(&sql, out, where)
	return c
}

func (c *CacheDB) Find(out interface{}, where ...interface{}) *CacheDB {
	sql := gorm.Statement{}
	key := ""
	key = cacheKey(sql)
	err := c.store.Get(key, out)
	if cache.IsNotHave(err) {
		log.Debugf("find no cache data")
		db := c.DB.Find(out, where)
		if err := db.Error; err == nil {
			c.store.Set(key, out, c.Expire)
			return c
		}
	}
	c.DB = c.DB.Find(&sql, out, where)
	return c
}

func (c *CacheDB) Count(count interface{}) *CacheDB {
	sql := gorm.Statement{}
	key := ""
	out := count.(int64)
	key = cacheKey(sql)
	err := c.store.Get(key, out)
	if cache.IsNotHave(err) {
		log.Debugf("count no cache data, err:%s", err)
		db := c.DB.Count(&out)
		if err := db.Error; err == nil {
			var value interface{}
			if v := reflect.ValueOf(&out); v.Kind() == reflect.Ptr {
				p := v.Elem()
				switch p.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					value = conv.IntPtrTo64(&out)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					value = conv.UintPtrTo64(&out)
				}
			}
			if err := c.store.Set(key, value, c.Expire); err != nil {
				c.DB.AddError(err)
			}
			return c
		}
	}
	c.DB = c.DB.Count(&out)
	return c
}

func cacheKey(gSql gorm.Statement) string {
	// sqlStr := fmt.Sprintf(CacheKeyFormat, sql.SQL, sql.SQLVars)
	sqlStr := sql.SqlParse(gSql.SQL.String(), gSql.Vars)
	return crypt.MD5([]byte(sqlStr))
}
