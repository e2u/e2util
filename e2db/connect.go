package e2db

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"gorm.io/driver/sqlite"
	gormlogger "gorm.io/gorm/logger"

	"github.com/e2u/e2util/e2model"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect 数据库连接
type Connect struct {
	*Config
	db   *gorm.DB
	roDb []*gorm.DB
}

type Option struct {
	Debug bool
}

type Config struct {
	Dialector         gorm.Dialector
	GormConfig        *gorm.Config
	Writer            string
	Reader            []string
	LogLevel          gormlogger.LogLevel
	Driver            string
	DisableAutoReport bool
}

func New(config *Config) *Connect {
	var err error
	var primaryDialector gorm.Dialector
	var slaveDialectors []gorm.Dialector
	conn := &Connect{
		Config: config,
	}
	log := NewLogger()
	log.LogLevel = config.LogLevel

	if config.GormConfig == nil {
		config.GormConfig = &gorm.Config{
			Logger: log,
		}
	}

	switch config.Driver {
	case "postgres", "postgresql", "pgsql":
		config.Dialector = postgres.Dialector{}
	case "mysql":
		config.Dialector = mysql.Dialector{}
	}

	if config.Dialector == nil {
		switch {
		case strings.Contains(config.Writer, "host="):
			config.Dialector = postgres.Dialector{}
		case strings.Contains(config.Writer, "@tcp("):
			config.Dialector = mysql.Dialector{}
		}
	}

	switch config.Dialector.Name() {
	case "mysql":
		if config.Writer != "" {
			primaryDialector = mysql.Open(config.Writer)
		}
		for _, dns := range config.Reader {
			slaveDialectors = append(slaveDialectors, mysql.Open(dns))
		}
	case "postgres":
		if config.Writer != "" {
			primaryDialector = postgres.Open(config.Writer)
		}
		for _, dns := range config.Reader {
			slaveDialectors = append(slaveDialectors, postgres.Open(dns))
		}
	case "sqlite":
		if config.Writer != "" {
			primaryDialector = sqlite.Open(config.Writer)
		}
		for _, dns := range config.Reader {
			slaveDialectors = append(slaveDialectors, sqlite.Open(dns))
		}
	}

	conn.db, err = gorm.Open(primaryDialector, config.GormConfig)
	if err != nil {
		logrus.Errorf("open primary connection error=%v", err)
		panic(err)
	}

	if config.Dialector.Name() == "sqlite" {
		conn.roDb = append(conn.roDb, conn.db)
	} else {
		for _, sd := range slaveDialectors {
			c, err := gorm.Open(sd, config.GormConfig)
			if err != nil {
				logrus.Errorf("open slave connection error=%v", err)
				continue
			}
			conn.roDb = append(conn.roDb, c)
		}
		if len(slaveDialectors) > 0 && len(conn.roDb) == 0 {
			panic(fmt.Errorf("no any slave connections"))
		}
	}

	return conn
}

func (c *Connect) RW(opts ...*Option) *gorm.DB {
	o := &Option{}
	if len(opts) > 0 {
		o = opts[0]
	}

	if o.Debug {
		return c.db.Debug()
	}
	return c.db
}

func (c *Connect) RO(opts ...*Option) *gorm.DB {
	if len(c.roDb) == 0 {
		logrus.Errorf("no read-only database connections")
		return nil
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(c.roDb))))

	o := &Option{}
	if len(opts) > 0 {
		o = opts[0]
	}

	if o.Debug {
		return c.roDb[n.Int64()].Debug()
	}

	return c.roDb[n.Int64()]
}

// Exists
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

func (c *Connect) Save(v interface{}) error {
	return c.RW().Save(v).Error
}

func (c *Connect) Delete(v interface{}) error {
	return c.RW().Delete(v).Error
}

func (c *Connect) Patch(v interface{}, patchs []*e2model.HttpPatch) error {
	updates := make(map[string]interface{})
	for _, patch := range patchs {
		updates[patch.Path] = patch.Value
	}
	return c.RW().Model(v).Updates(updates).Error
}

func (c *Connect) DebugRW() *gorm.DB {
	return c.RW().Debug()
}

func (c *Connect) DebugRO() *gorm.DB {
	return c.RO().Debug()
}
