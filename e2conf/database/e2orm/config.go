package e2orm

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Config 数据库连接配置
type Config struct {
	Endpoint    string   // 读写
	RoEndpoint  []string // 只读,可设置多个
	Active      int
	Idle        int
	IdleTimeout time.Duration
}

// Options Dao 操作的選項
type Options struct {
	Unscoped bool
}

func (o *Options) parsed(db *gorm.DB) *gorm.DB {
	if o.Unscoped {
		db = db.Unscoped()
	}
	return db
}

func opt(opts ...*Options) *Options {
	if len(opts) == 0 {
		return &Options{}
	}
	return opts[0]
}
