package e2conf

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/e2u/e2util/e2conf/cache/e2redis"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/e2u/e2util/e2conf/e2http"
	"github.com/e2u/e2util/e2conf/logger/e2logrus"
	"github.com/e2u/e2util/e2env"
	"github.com/spf13/viper"
)

type Config struct {
	Env    string
	Http   *e2http.Config
	Orm    *e2db.Config
	Redis  *e2redis.Config
	Logger *e2logrus.Config
	// General *e2general.Config // 这里存的 key 都会转小写字母
	Viper *viper.Viper
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

func New(input *InitConfigInput) *Config {
	e2env.EnvStringVar(&env, "env", "dev", "application run env=[dev|sit|uat|prod|unit-test|...]")
	e2env.EnvStringVar(&logLevel, "log-level", "debug", "set logger level: [debug|info|warn|error]")
	e2env.EnvStringVar(&searchPath, "search-path", "", "set config search path")
	e2env.EnvStringVar(&configFile, "config", "", "set config file name")

	if !flag.Parsed() {
		flag.Parse()
	}

	if input != nil {
		if len(input.Env) == 0 {
			input.Env = env
		}
		if len(input.AddConfigPath) == 0 {
			input.AddConfigPath = []string{".", "./etc", "./conf", "./config", "./cfg"}
		}
		if len(input.ConfigName) == 0 {
			input.ConfigName = "app-" + env
		}
	} else {
		input = &InitConfigInput{
			Env:           env,
			AddConfigPath: []string{".", "./etc", "./conf", "./config", "./cfg"},
			ConfigName:    "app-" + env,
		}
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
		v.SetConfigType("toml")
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
	switch appConf.Logger.LogLevel {
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
	switch appConf.Logger.LogFormat {
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

	if len(appConf.Logger.LogLevel) == 0 {
		appConf.Logger.LogLevel = "debug"
	}

	ll, err := logrus.ParseLevel(appConf.Logger.LogLevel)
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
	//if len(v.GetStringMap("general")) > 0 {
	//	appConf.General.PutAll(v.GetStringMap("general"))
	//}
}
