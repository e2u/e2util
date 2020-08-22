package e2conf

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/e2u/e2util/e2conf/e2general"
	"github.com/e2u/e2util/e2conf/e2http"
	"github.com/e2u/e2util/e2conf/logger/e2logrus"
	"github.com/e2u/e2util/e2env"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Env     string
	Http    *e2http.Config
	Logger  *e2logrus.Config
	General *e2general.Config // 这里存的 key 都会转小写字母
}

var (
	env        string
	configPath string
)

func New() *Config {
	cfg := &Config{
		General: e2general.New(),
	}

	// 如果是单元测试，则不加载任何配置文件
	if strings.ToLower(os.Getenv("DEV_UNIT_TEST")) == "true" {
		return cfg
	}

	e2env.EnvStringVar(&env, "env", "dev", "application run env=[dev|sit|uat|prod]")
	e2env.EnvStringVar(&configPath, "config-path", ".", "set config search path")
	flag.Parse()
	v := viper.New()
	v.SetConfigName("app-" + env)
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.AddConfigPath("./etc")
	v.AddConfigPath("./conf")
	v.AddConfigPath("./config")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	cfg.Env = env
	if len(v.GetStringMap("general")) > 0 {
		cfg.General.PutAll(v.GetStringMap("general"))
	}

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

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	{
		if cfg.Logger != nil && len(cfg.Logger.LogLevel) == 0 {
			cfg.Logger.LogLevel = "debug"
		}
		ll, err := logrus.ParseLevel(cfg.Logger.LogLevel)
		if err != nil {
			ll = logrus.DebugLevel
		}
		logrus.SetLevel(ll)
	}

	rl, err := e2logrus.NewWriter(cfg.Logger)
	if err != nil {
		panic(err.Error())
	}
	logrus.SetOutput(rl)

	logrus.Trace("logrus trace level active")
	logrus.Debug("logrus debug level active")
	logrus.Info("logrus info level active")
	logrus.Warnf("logrus warn level active")
	logrus.Error("logrus error level active")

	return cfg
}
