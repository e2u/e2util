package e2db

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/e2u/e2util/e2model"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connect struct {
	*Config
	db        *gorm.DB
	roDb      []*gorm.DB
	dialector gorm.Dialector
}

type Option struct {
	Debug bool
}

type Config struct {
	*gorm.Config
	Writer            string   `mapstructure:"writer"`
	Reader            []string `mapstructure:"reader"`
	DBLogLevel        string   `mapstructure:"log_level"`
	LogAdapter        string   `mapstructure:"log_adapter"`
	Driver            string   `mapstructure:"driver"`
	DisableAutoReport bool     `mapstructure:"disable_auto_report"`
	EnableDebug       bool     `mapstructure:"enable_debug"`
}

func New(cfg *Config) *Connect {
	var err error
	var primaryDialector gorm.Dialector
	var slaveDialector []gorm.Dialector
	conn := &Connect{
		Config: cfg,
	}
	log := NewLogger(cfg.DBLogLevel, cfg.LogAdapter)

	if cfg.Config == nil {
		cfg.Config = &gorm.Config{
			Logger: log,
		}
	}

	switch cfg.Driver {
	case "postgres", "postgresql", "pgsql":
		conn.dialector = postgres.Dialector{}
	case "mysql":
		conn.dialector = mysql.Dialector{}
	case "sqlite", "sqlite3":
		conn.dialector = sqlite.Dialector{}
	case "go-sqlite":
		conn.dialector = sqlite.Dialector{}
	}

	if conn.dialector == nil {
		switch {
		case strings.Contains(cfg.Writer, "host="):
			conn.dialector = postgres.Dialector{}
		case strings.Contains(cfg.Writer, "@tcp("):
			conn.dialector = mysql.Dialector{}
		case strings.HasPrefix(cfg.Writer, "file:"):
			conn.dialector = sqlite.Dialector{}
		}
	}

	switch conn.dialector.Name() {
	case "mysql":
		// user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
		if cfg.Writer != "" {
			primaryDialector = mysql.Open(cfg.Writer)
		}
		for _, dns := range cfg.Reader {
			slaveDialector = append(slaveDialector, mysql.Open(dns))
		}

	case "postgres":
		// host=127.0.0.1 port=5432 user=postgres password=none dbname=db1 sslmode=disable application_name=apa01
		if cfg.Writer != "" {
			primaryDialector = postgres.Open(cfg.Writer)
		}
		for _, dns := range cfg.Reader {
			slaveDialector = append(slaveDialector, postgres.Open(dns))
		}
	case "sqlite", "go-sqlite":
		// file:db1?mode=memory&cache=shared
		if cfg.Writer != "" {
			primaryDialector = sqlite.Open(cfg.Writer)
		}
		for _, dns := range cfg.Reader {
			slaveDialector = append(slaveDialector, sqlite.Open(dns))
		}
	}

	conn.db, err = gorm.Open(primaryDialector, cfg.Config)
	if err != nil {
		logrus.Errorf("open primary connection error=%v", err)
		panic(err)
	}

	if conn.dialector.Name() == "sqlite" {
		conn.roDb = append(conn.roDb, conn.db)
	} else {
		for _, sd := range slaveDialector {
			c, err := gorm.Open(sd, cfg.Config)
			if err != nil {
				logrus.Errorf("open slave connection error=%v", err)
				continue
			}
			conn.roDb = append(conn.roDb, c)
		}
		if len(slaveDialector) > 0 && len(conn.roDb) == 0 {
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
	if o.Debug || c.EnableDebug {
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

	if o.Debug || c.EnableDebug {
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

func (c *Connect) AutoMigrate(dst ...interface{}) {
	if err := c.DebugRW().AutoMigrate(dst...); err != nil {
		slog.Error("gorm auto migrate model error", "error", err, "model", dst)
	}
}

func (c *Connect) CreateSchema(schemas ...string) {
	for _, schema := range schemas {
		if err := c.DebugRW().Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error; err != nil {
			slog.Error("gorm create schema error", "error", err, "schema", schema)
		}
	}
}
