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
	AppName string
	Env     string
	DB      *e2db.Connect
	Cache   *e2cache.Connect
	Http    *e2http.Config
}
type DefaultConfig struct {
	Orm    *e2db.Config     `mapstructure:"orm"`
	Http   *e2http.Config   `mapstructure:"http"`
	Logger *e2logrus.Config `mapstructure:"logger"`
	Cache  *e2cache.Config  `mapstructure:"cache"`
}

func parseEnvAndFlags() {
	viper.AutomaticEnv()
	viper.SetConfigType("toml")
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
var configFS embed.FS

func SetConfigFS(fs embed.FS) {
	configFS = fs
}
func New(ctx context.Context) *Context {
	var c *Context
	once.Do(func() {
		c = newContext(ctx)
	})
	return c
}

func newContext(ctx context.Context) *Context {
	rc := &Context{
		Context: ctx,
	}
	viper.SetDefault("env", "dev")
	viper.SetConfigType("toml")

	parseEnvAndFlags()

	rc.Env = viper.GetString("env")
	rc.AppName = viper.GetString("app-name")

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

	if cfg.Orm != nil {
		rc.DB = e2db.New(cfg.Orm)
	}

	if cfg.Cache != nil {
		if cfg.Cache.Enable {
			rc.Cache = e2cache.New(cfg.Cache)
		} else {
			cfg.Cache.Type = "fake"
			rc.Cache = e2cache.New(cfg.Cache)
		}
	}

	if cfg.Http != nil {
		rc.Http = cfg.Http
	}

	return rc
}

// func initLogger(cfg *e2logrus.Config) {
//	if cfg.Format == "json" {
//		logrus.SetFormatter(&logrus.JSONFormatter{
//			TimestampFormat:   time.RFC3339Nano,
//			DisableTimestamp:  cfg.DisableFullTimestamp,
//			DisableHTMLEscape: cfg.DisableHTMLEscape,
//			DataKey:           e2var.NeverDefault(cfg.DataKey, "fields"),
//			PrettyPrint:       cfg.PrettyPrint,
//		})
//	} else {
//		logrus.SetFormatter(&logrus.TextFormatter{
//			ForceColors:               !cfg.DisableColor, //
//			DisableColors:             cfg.DisableColor,  //
//			ForceQuote:                !cfg.DisableQuote,
//			DisableQuote:              cfg.DisableQuote,
//			EnvironmentOverrideColors: cfg.EnvironmentOverrideColors, //
//			DisableTimestamp:          cfg.DisableFullTimestamp,
//			FullTimestamp:             !cfg.DisableFullTimestamp,
//			TimestampFormat:           time.RFC3339Nano,
//			PadLevelText:              !cfg.DisablePadLevelText,
//			QuoteEmptyFields:          !cfg.DisableQuoteEmptyFields, //
//		})
//	}
//
//	logrus.SetReportCaller(!cfg.DisableReportCaller)
//
//	rotationTime := cfg.RotationTime
//	if rotationTime <= 0 {
//		rotationTime = 86400
//	}
//	maxAge := cfg.MaxAge
//	if maxAge <= 0 {
//		maxAge = 365
//	}
//
//	logLevelStr := cfg.Level
//	if logLevelStr == "" {
//		logLevelStr = logrus.InfoLevel.String()
//	}
//
//	logLevel, err := logrus.ParseLevel(logLevelStr)
//	if err != nil {
//		logLevel = logrus.InfoLevel
//	}
//	logrus.SetLevel(logLevel)
//
//	output := cfg.Output
//	if output == "" {
//		output = "stdout"
//	}
//	if strings.HasPrefix(output, "file://") {
//		logPath := output[7:]
//		linkName := filepath.Join(filepath.Dir(logPath), "current")
//		rl, err := rotatelogs.New(logPath,
//			rotatelogs.WithMaxAge(24*time.Hour*time.Duration(maxAge)),
//			rotatelogs.WithRotationTime(time.Second*time.Duration(rotationTime)),
//			rotatelogs.WithLinkName(linkName),
//		)
//		if err != nil {
//			logrus.Fatal(err)
//		}
//		logrus.SetOutput(rl)
//	} else {
//		logrus.SetOutput(os.Stdout)
//	}
//	logrus.AddHook(&e2logrus.SeqHook{})
//
//	logrus.Trace("active logrus TRACE level")
//	logrus.Debug("active logrus DEBUG level")
//	logrus.Info("active logrus INFO level")
//	logrus.Warn("active logrus WARN level")
//	logrus.Error("active logrus ERROR level")
//}
