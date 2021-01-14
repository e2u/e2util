package e2db

import (
	"math/rand"

	gormlogger "gorm.io/gorm/logger"

	"github.com/e2u/e2util/e2model"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect 数据库连接
type Connect struct {
	db   *gorm.DB
	roDb []*gorm.DB
}

// Config 數據庫鏈接配置
type Config struct {
	Dialector  gorm.Dialector
	GormConfig *gorm.Config
	PrimaryDns string
	SlaveDns   []string
	LogLevel   gormlogger.LogLevel
}

func New(config *Config) *Connect {
	var err error
	var primaryDialector gorm.Dialector
	var slaveDialectors []gorm.Dialector
	conn := &Connect{}
	log := NewLogger()
	log.LogLevel = config.LogLevel
	if config.GormConfig == nil {
		config.GormConfig = &gorm.Config{
			Logger: log,
		}
	}

	switch config.Dialector.Name() {
	case "mysql":
		primaryDialector = mysql.Open(config.PrimaryDns)
		for _, dns := range config.SlaveDns {
			slaveDialectors = append(slaveDialectors, mysql.Open(dns))
		}
	case "postgres":
		primaryDialector = postgres.Open(config.PrimaryDns)
		for _, dns := range config.SlaveDns {
			slaveDialectors = append(slaveDialectors, postgres.Open(dns))
		}
	case "sqlite":
		primaryDialector = sqlite.Open(config.PrimaryDns)
		for _, dns := range config.SlaveDns {
			slaveDialectors = append(slaveDialectors, sqlite.Open(dns))
		}
	}

	conn.db, err = gorm.Open(primaryDialector, config.GormConfig)
	if err != nil {
		logrus.Errorf("open primary connection error=%v", err)
	}

	for _, sd := range slaveDialectors {
		c, err := gorm.Open(sd, config.GormConfig)
		if err != nil {
			logrus.Errorf("open slave connection error=%v", err)
			continue
		}
		conn.roDb = append(conn.roDb, c)
	}
	return conn
}

// RW 返回主數據（讀寫）連接
func (c *Connect) RW() *gorm.DB {
	return c.db
}

// RO 返回從數據（只讀）連接
func (c *Connect) RO() *gorm.DB {
	return c.roDb[rand.Intn(len(c.roDb))]
}

// Exists 檢查記錄指定條件的記錄是否存在
// 從讀寫庫查找，避免數據同步問題
// 使用方法:
// exs := d.Exists(&model.Dictionary{}, "category_code = ? and code = ?", categoryCode, code)
// return exs.Bool, exs.Error
func (c *Connect) Exists(v interface{}, query string, where ...interface{}) *e2model.NullBool {
	var count int64
	if err := c.RW().Model(v).
		Where(query, where...).
		Count(&count).Error; err != nil {
		return e2model.NewNullBool(false, err)
	}
	return e2model.NewNullBool(count > 0, nil)
}
