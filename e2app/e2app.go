package e2app

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/e2u/e2util/e2cache"
	"github.com/e2u/e2util/e2db"
	"github.com/e2u/e2util/e2http"
	"github.com/e2u/e2util/e2logrus"
	"github.com/e2u/e2util/e2os"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Context struct {
	context.Context
	Env   string
	App   *AppConfig
	DB    *e2db.Connect
	Cache *e2cache.Connect
	Http  *e2http.Config
}

type DefaultConfig struct {
	App    *AppConfig       `mapstructure:"app"`
	Orm    *e2db.Config     `mapstructure:"orm"`
	Http   *e2http.Config   `mapstructure:"http"`
	Logger *e2logrus.Config `mapstructure:"logger"`
	Cache  *e2cache.Config  `mapstructure:"cache"`
}

func parseEnvAndFlags() {
	viper.AutomaticEnv()
	viper.SetConfigType("toml")
	viper.SetDefault("env", "dev")
	// viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	pflag.String("app-name", "", "Setting app name")
	pflag.String("env", "dev", "Setting the environment to use. [dev|test|prod]")
	pflag.String("log-level", "debug", "Setting the logger level: [debug|info|warn|error]")
	pflag.String("config", "", "Setting config path")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		logrus.Fatal(err)
		// 無法綁定命令行標誌: %v
	}
	if err := viper.BindEnv("env", "ENV"); err != nil {
		logrus.Fatal(err)
		// 無法綁定環境變量 env: %v
	}
	if err := viper.BindEnv("log-level", "LOG_LEVEL"); err != nil {
		logrus.Fatal(err)
		// 無法綁定環境變量 log-level: %v
	}
	if err := viper.BindEnv("config", "CONFIG"); err != nil {
		logrus.Fatal(err)
		// 無法綁定環境變量 config: %v
	}
	if err := viper.BindEnv("app-name", "APP_NAME"); err != nil {
		logrus.Fatal(err)
		// 無法綁定環境變量 app-name: %v
	}
}

var once sync.Once

func New(args ...any) *Context {
	var c *Context
	once.Do(func() {
		c = newContext(args...)
	})
	return c
}

func newContext(args ...any) *Context {
	var configFS embed.FS
	var ctx context.Context
	fileExt := ".toml" // Default extension
	for _, arg := range args {
		if v, ok := arg.(embed.FS); ok {
			configFS = v
		}
		if v, ok := arg.(context.Context); ok {
			ctx = v
		}
	}

	parseEnvAndFlags()

	rc := &Context{Context: ctx, Env: viper.GetString("env")}

	if cfgFile := viper.GetString("config"); cfgFile != "" {
		logrus.Info("Loading config file")
		if !e2os.FileExists(cfgFile) {
			logrus.Fatalf("config file %s does not exist", cfgFile)
		}
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			logrus.Fatalf("failed to read config file %s: %v", cfgFile, err)
		}
	} else if _, err := configFS.ReadDir("."); err == nil { // Non-empty FS
		logrus.Info("Loading config FS")
		if rc.Env == "" {
			logrus.Warn("rc.Env is empty, falling back to default paths")
		} else {
			configFile := rc.Env + fileExt
			configData, err := configFS.ReadFile(configFile)
			if err != nil {
				logrus.Warnf("embedded config %s not found: %v", configFile, err)
			} else {
				if err := viper.ReadConfig(bytes.NewReader(configData)); err != nil {
					logrus.Fatalf("failed to read embedded config %s: %v", configFile, err)
				}
				goto configLoaded
			}
		}
	}

	// Fallback to default paths
	logrus.Info("Loading config file in paths")
	for _, ap := range []string{".", "./etc", "./conf", "./config", "./cfg"} {
		viper.AddConfigPath(ap)
	}
	viper.SetConfigName(strings.TrimSuffix(rc.Env, fileExt))
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("failed to read config in paths: %v", err)
	}

configLoaded:
	cfg := &DefaultConfig{}

	if err := viper.Unmarshal(cfg); err != nil {
		logrus.Fatalf("failed to unmarshal config: %v", err)
	}

	if cfg.Logger != nil {
		logrus.Info("Logger begin configuration")
		if l := e2logrus.NewLogger(cfg.Logger); l != nil {
			logrus.SetFormatter(l.Formatter)
			logrus.SetOutput(l.Out)
			logrus.SetReportCaller(l.ReportCaller)
			logrus.SetLevel(l.Level)
		}
		logrus.Info("Logger configured")
	}

	if cfg.App != nil {
		rc.App = cfg.App
	} else {
		rc.App = &AppConfig{}
	}

	if v := viper.GetString("app-name"); v != "" {
		rc.App.Name = v
	}

	if cfg.Orm != nil {
		logrus.Info("Orm begin configuration")
		if db, err := e2db.New(cfg.Orm); err != nil {
			logrus.Fatalf("failed to initialize DB: %v", err)
		} else {
			rc.DB = db
		}
		logrus.Info("Orm configured")
	}

	if cfg.Cache != nil {
		logrus.Info("Cache begin configuration")
		if cfg.Cache.Enable {
			rc.Cache = e2cache.New(cfg.Cache)
		} else {
			logrus.Info("Cache is disabled")
			cfg.Cache.Type = "fake"
			rc.Cache = e2cache.New(cfg.Cache)
		}
		logrus.Info("Cache configured")
	} else {
		rc.Cache = e2cache.New(&e2cache.Config{
			Enable: false,
			Type:   "fake",
		})
		logrus.Info("Initializing disabled fake cache")
	}

	if cfg.Http != nil {
		rc.Http = cfg.Http
	}

	logrus.Info("Using config:", viper.ConfigFileUsed())
	logrus.Info("Running Env:", rc.Env)

	return rc
}

func (c *Context) ExtendedConfig(e any) error {
	return viper.Unmarshal(e)
}

func DebugFS(embeddedFS embed.FS, dir string) error {
	// Read directory entries
	fmt.Println("DebugFS beginning")
	entries, err := embeddedFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	// Iterate through entries
	for _, entry := range entries {
		// Construct full path
		fullPath := path.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively list files in subdirectory
			if err := DebugFS(embeddedFS, fullPath); err != nil {
				return err
			}
		} else {
			// Print file path
			fmt.Println(fullPath)
		}
	}

	return nil
}
