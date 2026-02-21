package e2db

import (
	"context"
	"errors"
	"fmt"
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

type Option struct {
	Debug bool
}

func (c *Connect) Exists(v any, query string, useRO bool, where ...any) *e2model.NullBool {
	db := c.RW()
	if useRO {
		db = c.RO()
	}
	var count int64
	if err := db.Model(v).Where(query, where...).Count(&count).Error; err != nil {
		return e2model.NewNullBool(false, err)
	}
	return e2model.NewNullBool(count > 0, nil)
}

func (c *Connect) Patch(ctx context.Context, v any, patchs []*e2model.HttpPatch) error {
	updates := make(map[string]any)
	for _, patch := range patchs {
		if patch.Path == "" {
			return fmt.Errorf("invalid patch: empty path")
		}
		updates[patch.Path] = patch.Value
	}
	return c.RW().WithContext(ctx).Model(v).Updates(updates).Error
}

func (c *Connect) AutoMigrate(ctx context.Context, dst ...any) error {
	if !c.EnableMigrate {
		logrus.Warn("migrate disabled")
		return nil
	}

	err := c.RW().WithContext(ctx).AutoMigrate(dst...)
	if err != nil {
		logrus.Errorf("gorm auto migrate model error=%v, model=%v", err, dst)
		return fmt.Errorf("auto migrate failed: %w", err)
	}
	return nil
}

func (c *Connect) CreateUniqueIndexWithNulls(ctx context.Context, tableName string, indexName string, dropExists bool, columns ...string) error {
	if !c.EnableMigrate {
		logrus.Warn("migrate disabled")
		return nil
	}

	indexName = "uidx_" + tableName + "_" + indexName
	if dropExists {
		if err := c.RW().WithContext(ctx).Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName)).Error; err != nil {
			return err
		}
	}
	return c.RW().WithContext(ctx).Exec(fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s) NULLS NOT DISTINCT;", indexName, tableName, strings.Join(columns, ","))).Error
}

func (c *Connect) CreateSchema(ctx context.Context, schemas ...string) error {
	if c.dialector.Name() != "postgres" {
		return fmt.Errorf("CreateSchema is only supported for PostgreSQL")
	}
	for _, schema := range schemas {
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(schema) {
			return fmt.Errorf("invalid schema name: %s", schema)
		}
		err := c.RW().WithContext(ctx).Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)).Error
		if err != nil {
			logrus.Errorf("gorm create schema error=%v, schema=%s", err, schema)
			return fmt.Errorf("failed to create schema %s: %w", schema, err)
		}
	}
	return nil
}

func createPostgresDatabase(dsn string) error {
	var dsnc []string
	var dbname string
	for d := range strings.SplitSeq(dsn, " ") {
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

// --------------------------------

// Config defines the database connection configuration.
type Config struct {
	*gorm.Config
	Writer                          string           `mapstructure:"writer"`
	Readers                         []string         `mapstructure:"readers"`
	Driver                          string           `mapstructure:"driver"`
	DisableAutoReport               bool             `mapstructure:"disable_auto_report"`
	EnableDebug                     bool             `mapstructure:"enable_debug"`
	EnableMigrate                   bool             `mapstructure:"enable_migrate"`
	AutoCreateDatabase              bool             `mapstructure:"auto_create_database"`
	InitSqls                        []string         `mapstructure:"init_sqls"`
	SQLLogSlowThreshold             int              `mapstructure:"sql_log_slow_threshold"`
	SQLLogIgnoreRecordNotFoundError bool             `mapstructure:"sql_log_ignore_record_not_found_error"`
	SQLLogColorful                  bool             `mapstructure:"sql_log_colorful"`
	LoggerConfig                    *e2logrus.Config `mapstructure:"logger"`
}

// Connect manages database connections with read-write separation.
type Connect struct {
	*Config
	db        *gorm.DB
	roDb      []*gorm.DB
	dialector gorm.Dialector
	roIndex   int // For round-robin selection
}

// New creates a new database connection instance.
func New(cfg *Config) (*Connect, error) {
	if cfg.Config == nil {
		cfg.Config = &gorm.Config{Logger: newLogger(cfg)}
	}

	conn := &Connect{Config: cfg}
	driver := strings.ToLower(cfg.Driver)

	// Detect and set dialector
	dialector, err := detectDialector(driver, cfg.Writer)
	if err != nil {
		return nil, fmt.Errorf("invalid driver or DSN: %w", err)
	}
	conn.dialector = dialector

	// Open primary (writer) database connection
	var primaryDialector gorm.Dialector
	switch driver {
	case "mysql":
		primaryDialector = mysql.Open(cfg.Writer)
	case "postgres", "postgresql", "pgsql":
		primaryDialector = postgres.Open(cfg.Writer)
	case "sqlite", "sqlite3", "go-sqlite":
		primaryDialector = sqlite.Open(cfg.Writer)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := gorm.Open(primaryDialector, cfg.Config)
	if err != nil {
		if cfg.AutoCreateDatabase {
			if err := ensureDatabaseExists(driver, cfg.Writer); err != nil {
				return nil, fmt.Errorf("failed to create database: %w", err)
			}
			db, err = gorm.Open(primaryDialector, cfg.Config)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to open primary DB: %w", err)
		}
	}
	conn.db = db

	// Execute initialization SQLs
	for _, sql := range cfg.InitSqls {
		if err := db.Exec(sql).Error; err != nil {
			logrus.Warnf("failed to execute init SQL '%s': %v", sql, err)
		}
	}

	// Set up read-only connections
	if driver == "sqlite" {
		conn.roDb = append(conn.roDb, db)
	} else {
		for _, dsn := range cfg.Readers {
			var roDialector gorm.Dialector
			switch driver {
			case "mysql":
				roDialector = mysql.Open(dsn)
			case "postgres", "postgresql", "pgsql":
				roDialector = postgres.Open(dsn)
			case "sqlite", "sqlite3", "go-sqlite":
				roDialector = sqlite.Open(dsn)
			}
			roDB, err := gorm.Open(roDialector, cfg.Config)
			if err != nil {
				logrus.Errorf("failed to open read-only DB '%s': %v", dsn, err)
				continue
			}
			conn.roDb = append(conn.roDb, roDB)
		}
		if len(cfg.Readers) > 0 && len(conn.roDb) == 0 {
			return nil, errors.New("no valid read-only connections available")
		}
	}

	return conn, nil
}

// RO returns a read-only database connection using round-robin selection.
func (c *Connect) RO(opts ...*Option) *gorm.DB {
	if len(c.roDb) == 0 {
		logrus.Errorf("no read-only database connections available, falling back to RW")
		return c.RW(opts...)
	}
	c.roIndex = (c.roIndex + 1) % len(c.roDb)
	db := c.roDb[c.roIndex]
	if len(opts) > 0 && (opts[0].Debug || c.EnableDebug) {
		return db.Debug()
	}
	return db
}

// RW returns the read-write database connection.
func (c *Connect) RW(opts ...*Option) *gorm.DB {
	db := c.db
	if len(opts) > 0 && (opts[0].Debug || c.EnableDebug) {
		return db.Debug()
	}
	return db
}

// detectDialector infers the dialector based on driver or DSN.
func detectDialector(driver, dsn string) (gorm.Dialector, error) {
	switch strings.ToLower(driver) {
	case "mysql":
		return mysql.Dialector{}, nil
	case "postgres", "postgresql", "pgsql":
		return postgres.Dialector{}, nil
	case "sqlite", "sqlite3", "go-sqlite":
		return sqlite.Dialector{}, nil
	default:
		if strings.Contains(dsn, "@tcp(") {
			return mysql.Dialector{}, nil
		}
		if strings.Contains(dsn, "host=") {
			return postgres.Dialector{}, nil
		}
		if strings.HasPrefix(dsn, "file:") {
			return sqlite.Dialector{}, nil
		}
		return nil, fmt.Errorf("unknown driver or DSN format: %s", dsn)
	}
}

// ensureDatabaseExists creates the database if it doesn't exist.
func ensureDatabaseExists(driver, dsn string) error {
	switch driver {
	case "mysql":
		return createMySQLDatabase(dsn)
	case "postgres", "postgresql", "pgsql":
		return createPostgresDatabase(dsn)
	default:
		return nil // SQLite does not require database creation
	}
}
