package e2app

import (
	"bytes"
	"context"
	"embed"
	"reflect"
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
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	pflag.String("app-name", "", "Setting app name")
	pflag.String("env", "dev", "Setting the environment to use. [dev|test|prod]")
	pflag.String("log-level", "debug", "Setting the logger level: [debug|info|warn|error]")
	pflag.String("config", "", "Setting config  path")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		logrus.Fatal(err)
	}
	if err := viper.BindEnv("env", "ENV"); err != nil {
		logrus.Fatal(err)
	}
	if err := viper.BindEnv("log-level", "LOG_LEVEL"); err != nil {
		logrus.Fatal(err)
	}
	if err := viper.BindEnv("config", "CONFIG"); err != nil {
		logrus.Fatal(err)
	}
	if err := viper.BindEnv("app-name", "APP_NAME"); err != nil {
		logrus.Fatal(err)
	}
}

var once sync.Once

func New(args ...any) *Context {
	var c *Context
	once.Do(func() {
		c = newContext(args)
	})
	return c
}

func newContext(args ...any) *Context {
	var configFS embed.FS
	var ctx context.Context
	for _, arg := range args {
		if v, ok := arg.(embed.FS); ok && !reflect.ValueOf(arg).IsNil() {
			configFS = v
		}
		if v, ok := arg.(context.Context); ok && !reflect.ValueOf(arg).IsNil() {
			ctx = v
		}
	}
	parseEnvAndFlags()

	rc := &Context{
		Context: ctx,
		Env:     viper.GetString("env"),
	}

	if cfgFile := viper.GetString("config"); cfgFile != "" {
		logrus.Info("Loading config file")
		if !e2os.FileExists(cfgFile) {
			logrus.Fatalf("config file %s does not exist", cfgFile)
		}
		viper.SetConfigFile(cfgFile)
	} else if !reflect.DeepEqual(configFS, embed.FS{}) {
		logrus.Info("Loading config FS")
		configData, err := configFS.ReadFile(rc.Env + ".toml")
		if err != nil {
			logrus.Fatal(err)
		}
		if err := viper.ReadConfig(bytes.NewReader(configData)); err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Info("Loading config file in paths")
		for _, ap := range []string{".", "./etc", "./conf", "./config", "./cfg"} {
			viper.AddConfigPath(ap)
		}
		viper.SetConfigName(rc.Env)
		if err := viper.ReadInConfig(); err != nil {
			logrus.Fatal(err)
		}
	}

	cfg := &DefaultConfig{}

	if err := viper.Unmarshal(cfg); err != nil {
		logrus.Fatal(err)
	}

	if cfg.Logger != nil {
		if l := e2logrus.NewLogger(cfg.Logger); l != nil {
			logrus.SetFormatter(l.Formatter)
			logrus.SetOutput(l.Out)
			logrus.SetReportCaller(l.ReportCaller)
			logrus.SetLevel(l.Level)
		}
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
		rc.DB = e2db.New(cfg.Orm)
	}

	if cfg.Cache != nil && cfg.Cache.Enable {
		rc.Cache = e2cache.New(cfg.Cache)
	} else {
		cfg.Cache.Type = "fake"
		rc.Cache = e2cache.New(cfg.Cache)
	}

	if cfg.Http != nil {
		rc.Http = cfg.Http
	}

	return rc
}
