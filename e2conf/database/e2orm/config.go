package e2orm

import (
	"time"
)

// Config 数据库连接配置
type Config struct {
	Endpoint    string   // 读写
	RoEndpoint  []string // 只读,可设置多个
	Active      int
	Idle        int
	IdleTimeout time.Duration
}
