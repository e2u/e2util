package e2conf

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/e2u/e2util/e2conf/cache/e2redis"
	"github.com/e2u/e2util/e2conf/e2http"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/e2u/e2util/e2conf/logger/e2logrus"
	"github.com/e2u/e2util/e2env"
	"github.com/spf13/viper"
)

type Config struct {
	Env    string
	Http   *e2http.Config   `mapstructure:"http"`
	Orm    *e2db.Config     `mapstructure:"orm"`
	Redis  *e2redis.Config  `mapstructure:"redis"`
	Logger *e2logrus.Config `mapstructure:"logger"`
	Viper  *viper.Viper     `mapstructure:"-"`
}

var (
	v          = viper.New()
	env        string
	searchPath string
	logLevel   string
	configFile string
	appConf    *Config
)

type InitConfigInput struct {
	Env           string
	ConfigFs      embed.FS
	AddConfigPath []string
	ConfigName    string
}

func init() {
	info := `
Important Change of the config
****Remove these from the Config struct: 
Http
Orm
Redis


[logger] parse by e2logrus.Config
1) LogLevel => level
2) LogFormat => format
3) AddSource => add_source 
4) MaxAge => max_age


[orm] parse by e2db.Connect
1) DBLogLevel => log_level
2) LogAdapter => log_adapter
3) Driver => driver
4) remove EnableTxDB
5) DisableAutoReport => disable_auto_report
6) EnableDebug => enable_debug
7) Writer => writer
8) Reader => reader

[redis] parse by e2redis.Config
1) Writer => writer
2) Reader => reader
`

	fmt.Println(info)
}

func New[T *InitConfigInput | embed.FS](args T) *Config {
	e2env.EnvStringVar(&env, "env", "dev", "application run env=[dev|sit|uat|prod|unit-test|...]")
	e2env.EnvStringVar(&logLevel, "log-level", "debug", "set logger level: [debug|info|warn|error]")
	e2env.EnvStringVar(&searchPath, "search-path", "", "set config search path")
	e2env.EnvStringVar(&configFile, "config", "", "set config file name")

	if !flag.Parsed() {
		flag.Parse()
	}

	defaultPath := []string{".", "./etc", "./conf", "./config", "./cfg"}

	input := &InitConfigInput{}

	switch v := any(args).(type) {
	case *InitConfigInput:
		input = v
	case embed.FS:
		input.ConfigFs = v
	}

	if len(input.Env) == 0 {
		input.Env = env
	}
	if len(input.AddConfigPath) == 0 {
		input.AddConfigPath = defaultPath
	}
	if len(input.ConfigName) == 0 {
		input.ConfigName = "app-" + env
	}

	appConf = &Config{
		Env: input.Env,
	}

	filename := input.ConfigName + ".toml"
	fmt.Printf("> env: %s\n", env)
	fmt.Printf("> env search-path: %s\n", searchPath)
	fmt.Printf("> env config: %s\n", configFile)
	fmt.Printf("> config file=%v\n", filename)

	// check embed fs path
	for _, ap := range input.AddConfigPath {
		f, err := input.ConfigFs.Open(filepath.Join(ap, filename))
		if err == nil {
			if fst, err := f.Stat(); err == nil && fst.Size() > 0 {
				filename = filepath.Join(ap, filename)
			}
			_ = f.Close()
		}
	}

	v.SetConfigType("toml")

	if f, err := input.ConfigFs.Open(filename); err == nil {
		fmt.Printf("> load from embed fs, file name=%v\n", filename)
		defer func(f fs.File) {
			_ = f.Close()
		}(f)
		v.SetConfigFile(filename)
		if err := v.ReadConfig(f); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	} else {
		v.SetConfigName(input.ConfigName)
		for _, ap := range input.AddConfigPath {
			fmt.Printf("add config path: %v\n", ap)
			v.AddConfigPath(ap)
		}
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	unmarshalAppConfig(v)

	setupGin()
	setupLogrus()
	setupSlog()
	return appConf
}

func setupGin() {
	fmt.Println("> gin setup")
	switch env {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	case "dev":
		gin.SetMode(gin.DebugMode)
	case "sit", "uat":
		gin.DisableConsoleColor()
		gin.SetMode(gin.DebugMode)
	}
}

func setupSlog() {
	fmt.Println("> slog setup")
	if appConf.Logger == nil {
		appConf.Logger = e2logrus.DefaultConfig()
	}

	ll := slog.LevelDebug
	switch appConf.Logger.Level {
	case "trace", "debug":
		ll = slog.LevelDebug
	case "info":
		ll = slog.LevelInfo
	case "warn":
		ll = slog.LevelWarn
	case "error":
		ll = slog.LevelError
	}
	opt := &slog.HandlerOptions{
		AddSource: appConf.Logger.AddSource,
		Level:     ll,
	}
	w, err := e2logrus.NewWriter(appConf.Logger)
	if err != nil {
		panic(err.Error())
	}

	var h slog.Handler
	switch appConf.Logger.Format {
	case "json":
		h = slog.NewJSONHandler(w, opt)
	default:
		h = slog.NewTextHandler(w, opt)
	}
	slog.SetDefault(slog.New(h))
	slog.Debug("active slog DEBUG level")
	slog.Info("active slog INFO level")
	slog.Warn("active slog WARN level")
	slog.Error("active slog ERROR level")
}

func setupLogrus() {
	fmt.Println("> logrus setup")
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableQuote:     true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})

	if appConf.Logger == nil {
		appConf.Logger = e2logrus.DefaultConfig()
	}

	if len(appConf.Logger.Level) == 0 {
		appConf.Logger.Level = "debug"
	}

	ll, err := logrus.ParseLevel(appConf.Logger.Level)
	if err != nil {
		ll = logrus.DebugLevel
	}
	logrus.SetLevel(ll)

	rl, err := e2logrus.NewWriter(appConf.Logger)
	if err != nil {
		panic(err.Error())
	}
	logrus.SetOutput(rl)

	logrus.Trace("active logrus TRACE level")
	logrus.Debug("active logrus DEBUG level")
	logrus.Info("active logrus INFO level")
	logrus.Warn("active logrus WARN level")
	logrus.Error("active logrus ERROR level")
}

func unmarshalAppConfig(v *viper.Viper) {
	appConf.Viper = v
	if err := v.Unmarshal(&appConf); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func (c *Config) GetStringMapStringByOS(key string) map[string]string {
	return getStringMapStringByOS(c.Viper, key)
}

func (c *Config) Unmarshal(key string, p any) error {
	return v.UnmarshalKey(key, &p)
}
