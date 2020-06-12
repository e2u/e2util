package e2general

import (
	"sync"

	"github.com/e2u/e2util/e2error"
)

// Config 通用配置
type Config struct {
	sync.RWMutex
	m map[string]interface{}
}

func New() *Config {
	return &Config{
		m: make(map[string]interface{}),
	}
}

// GetString 获取指定key的字符串值，如key不存在，返回错误
func (c *Config) GetString(key string) (string, error) {
	if v, ok := c.m[key]; ok {
		return v.(string), nil
	}
	return "", e2error.ErrIllegalParameter("key " + key + " not exists")
}

// GetStringDefault 获取指定key的字符串值，如key不存在，返回默认值
func (c *Config) GetStringDefault(key, defaultVal string) string {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if v, ok := c.m[key]; ok {
		return v.(string)
	}
	return defaultVal
}

func (c *Config) GetInt64(key string) (int64, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if v, ok := c.m[key]; ok {
		return v.(int64), nil
	}
	return 0, e2error.ErrIllegalParameter("key " + key + " not exists")
}

func (c *Config) GetInt64Default(key string, defaultVal int64) int64 {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if v, ok := c.m[key]; ok {
		return v.(int64)
	}
	return defaultVal
}

func (c *Config) GetBool(key string) (bool, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if v, ok := c.m[key]; ok {
		return v.(bool), nil
	}
	return false, e2error.ErrIllegalParameter("key " + key + " not exists")
}

func (c *Config) GetBoolDefault(key string, defaultVal bool) bool {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	if v, ok := c.m[key]; ok {
		return v.(bool)
	}
	return defaultVal
}

func (c *Config) Put(key string, value interface{}) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.m[key] = value
}

func (c *Config) PutAll(m map[string]interface{}) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.m = m
}
