package e2orm

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

// 連接緩存
var connCache sync.Map

// Connect 数据库连接
type Connect struct {
	db   *gorm.DB
	roDb []*gorm.DB
}

// DW 返回主(读写)连接
func (c *Connect) RW(opts ...*Options) *gorm.DB {
	if c.db != nil {
		return opt(opts...).parsed(c.db)
	}
	return nil
}

// RO 返回只读库连接, 若未设置只读连接，则返回主连接
func (c *Connect) RO(opts ...*Options) *gorm.DB {
	if len(c.roDb) == 0 {
		return opt(opts...).parsed(c.db)
	}
	return opt(opts...).parsed(c.roDb[rand.Intn(len(c.roDb))])
}

// Close 关闭数据库连接
func (c *Connect) Close() {
	if c.db != nil {
		_ = c.db.Close()
	}

	for _, cc := range c.roDb {
		if cc.DB() != nil {
			_ = cc.DB().Close()
		}
	}
}

// RWPing 主连接 ping
func (c *Connect) Ping(ctx context.Context) error {
	if c.RW().DB() == nil {
		return fmt.Errorf("rw disconnection")
	}
	if err := c.RW().DB().PingContext(ctx); err != nil {
		return fmt.Errorf("ping rw connection error,%v", err.Error())
	}

	var errMsg []string
	for _, cc := range c.roDb {
		if cc.DB() == nil {
			continue
		}
		if err := cc.DB().PingContext(ctx); err != nil {
			errMsg = append(errMsg, err.Error())
		}
	}
	if len(errMsg) > 0 {
		return fmt.Errorf("ping ro's connection error,%v", strings.Join(errMsg, ";"))
	}

	return nil
}

// newConnect 建立数据库连接
func newConnect(dialect string, c *Config) *Connect {
	cacheKey := func(endpoint string) string {
		return fmt.Sprintf("%s-%x", dialect, md5.Sum([]byte(endpoint)))
	}

	conn := func(e string) *gorm.DB {
		if v, ok := connCache.Load(cacheKey(e)); ok && v != nil && v.(*gorm.DB) != nil && v.(*gorm.DB).DB().Ping() == nil {
			return v.(*gorm.DB)
		}
		db, err := gorm.Open(dialect, e)
		if err != nil {
			panic(err)
		}
		db.DB().SetMaxIdleConns(c.Idle)
		db.DB().SetMaxOpenConns(c.Active)
		db.DB().SetConnMaxLifetime(c.IdleTimeout * time.Second)
		connCache.Store(cacheKey(e), db)
		return db
	}

	rc := &Connect{
		db: conn(c.Endpoint),
	}

	for _, e := range c.RoEndpoint {
		rc.roDb = append(rc.roDb, conn(e))
	}
	return rc
}
