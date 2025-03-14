package e2db

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/e2u/e2util/e2logrus"
	"github.com/e2u/e2util/e2model"
	"github.com/e2u/e2util/e2regexp"
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
	Writer                          string           `mapstructure:"writer"`
	Readers                         []string         `mapstructure:"readers"`
	Driver                          string           `mapstructure:"driver"`
	DisableAutoReport               bool             `mapstructure:"disable_auto_report"`
	EnableDebug                     bool             `mapstructure:"enable_debug"`
	AutoCreateDatabase              bool             `mapstructure:"auto_create_database"`
	InitSqls                        []string         `mapstructure:"init_sqls"`
	SQLLogSlowThreshold             int              `mapstructure:"sql_log_slow_threshold"`
	SQLLogIgnoreRecordNotFoundError bool             `mapstructure:"sql_log_ignore_record_not_found_error"`
	SQLLogColorful                  bool             `mapstructure:"sql_log_colorful"`
	LoggerConfig                    *e2logrus.Config `mapstructure:"logger"`
}

func New(cfg *Config) *Connect {
	var err error
	var primaryDialector gorm.Dialector
	var slaveDialector []gorm.Dialector

	if cfg.Config == nil {
		cfg.Config = &gorm.Config{}
	}

	cfg.Config.Logger = newLogger(cfg)

	conn := &Connect{
		Config: cfg,
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
		cfg.Driver = "mysql"
		if cfg.Writer != "" {
			primaryDialector = mysql.Open(cfg.Writer)
		}

		for _, dsn := range cfg.Readers {
			slaveDialector = append(slaveDialector, mysql.Open(dsn))
		}

	case "postgres":
		// host=127.0.0.1 port=5432 user=postgres password=none dbname=db1 sslmode=disable application_name=apa01
		cfg.Driver = "postgres"
		if cfg.Writer != "" {
			primaryDialector = postgres.Open(cfg.Writer)
		}

		for _, dsn := range cfg.Readers {
			slaveDialector = append(slaveDialector, postgres.Open(dsn))
		}
	case "sqlite", "go-sqlite":
		// file:db1?mode=memory&cache=shared
		cfg.Driver = "sqlite3"
		if cfg.Writer != "" {
			primaryDialector = sqlite.Open(cfg.Writer)
		}
		for _, dsn := range cfg.Readers {
			slaveDialector = append(slaveDialector, sqlite.Open(dsn))
		}
	}

	conn.db, err = gorm.Open(primaryDialector, cfg.Config)
	if err != nil {
		switch cfg.Driver {
		case "postgres":
			if cfg.AutoCreateDatabase && strings.Contains(err.Error(), "does not exist") {
				if err := createPostgresDatabase(cfg.Writer); err != nil {
					logrus.Fatalf("create postgres database error=%v", err)
					return nil
				}
			}
		case "mysql":
			if cfg.AutoCreateDatabase && strings.Contains(err.Error(), "Unknown database") {
				if err := createMySQLDatabase(cfg.Writer); err != nil {
					logrus.Fatalf("create mysql database error=%v", err)
					return nil
				}
			}
		}

		conn.db, err = gorm.Open(primaryDialector, cfg.Config)
		if err != nil {
			logrus.Panic(err)
		}
	}

	for _, s := range cfg.InitSqls {
		conn.db.Exec(s)
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
			logrus.Panic("no any slave connections")
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
// exs := d.exists(&model.Dictionary{}, "category_code = ? and code = ?", categoryCode, code)
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
		logrus.Errorf("gorm auto migrate model error=%v, model=%v", err, dst)
	}
}

func (c *Connect) CreateSchema(schemas ...string) {
	for _, schema := range schemas {
		if err := c.DebugRW().Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error; err != nil {
			logrus.Errorf("gorm create schema error=%v, model=%v", err, schema)
		}
	}
}

func createPostgresDatabase(dsn string) error {
	var dsnc []string
	var dbname string
	for _, d := range strings.Split(dsn, " ") {
		d = strings.TrimSpace(d)
		if strings.HasPrefix(strings.ToLower(d), "dbname=") {
			dbname, _ = strings.CutPrefix(d, "dbname=")
			continue
		}
		dsnc = append(dsnc, d)
	}
	if dbname == "" {
		return fmt.Errorf("database name is empty")
	}
	logrus.Infof("create new database %s at %s", dbname, strings.Join(dsnc, " "))
	tempDB, err := gorm.Open(postgres.Open(strings.Join(dsnc, " ")), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("failed to connect to Postgres: %v", err)
		return err
	}
	createSql := fmt.Sprintf(`CREATE DATABASE %s`, dbname)
	if err := tempDB.Exec(createSql).Error; err != nil {
		logrus.Fatalf("failed to create database: %v", err)
		return err
	}
	logrus.Info("Postgres database created successfully")
	return nil
}

func createMySQLDatabase(dsn string) error {
	re := regexp.MustCompile(`^(?P<userinfo>[^@]+)@(?P<conn>[^/]+)/(?P<dbname>[^\?]+)\?(?P<params>.+)$`)
	rs, ok := e2regexp.NamedFindStringSubmatch(dsn, re)
	if !ok {
		return fmt.Errorf("dsn parse error")
	}

	if _, ok := rs["dbname"]; !ok {
		return fmt.Errorf("dsn parse error")
	}

	tmpDSN := fmt.Sprintf("%s@%s/?%s", rs["userinfo"], rs["conn"], rs["params"])
	logrus.Infof("create new database %s at %s", rs["dbname"], tmpDSN)
	tempDB, err := gorm.Open(mysql.Open(tmpDSN), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("failed to connect to MySQL: %v", err)
		return err
	}
	createSql := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, rs["dbname"])
	if err := tempDB.Exec(createSql).Error; err != nil {
		logrus.Fatalf("failed to create database: %v", err)
		return err
	}
	logrus.Info("MySQL database created successfully")
	return nil
}
