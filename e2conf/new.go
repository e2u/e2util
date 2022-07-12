package e2conf

import (
	"embed"
	"flag"
	"fmt"

	"github.com/e2u/e2util/e2conf/cache/e2redis"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/e2u/e2util/e2conf/e2general"
	"github.com/e2u/e2util/e2conf/e2http"
	"github.com/e2u/e2util/e2conf/logger/e2logrus"
	"github.com/e2u/e2util/e2env"
	"github.com/spf13/viper"
)

type Config struct {
	Env     string
	Http    *e2http.Config
	Orm     *e2db.Config
	Orm2    *e2db.Config
	Orm3    *e2db.Config
	Orm4    *e2db.Config
	Redis   *e2redis.Config
	Logger  *e2logrus.Config
	General *e2general.Config // 这里存的 key 都会转小写字母
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
	fmt.Printf("> env: %s\n", env)
	fmt.Printf("> env search-path: %s\n", searchPath)
	fmt.Printf("> env config: %s\n", configFile)

	if input != nil {
		if len(input.Env) == 0 {
			input.Env = env
		}
		if len(input.AddConfigPath) == 0 {
			input.AddConfigPath = []string{".", "./etc", "./conf", "./config"}
		}
		if len(input.ConfigName) == 0 {
			input.ConfigName = "app-" + env
		}
	} else {
		input = &InitConfigInput{
			Env:           env,
			AddConfigPath: []string{".", "./etc", "./conf", "./config"},
			ConfigName:    "app-" + env,
		}
	}

	appConf = &Config{
		Env:     input.Env,
		General: e2general.New(),
	}

	filename := input.ConfigName + ".toml"
	fmt.Printf("> config file=%v\n", filename)

	if f, err := input.ConfigFs.Open(filename); err == nil {
		fmt.Printf("> load from embed fs, file name=%v\n", filename)
		defer f.Close()
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

func setupLogrus() {
	fmt.Println("> logrus setup")
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableQuote:     true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})

	if appConf.Logger == nil {
		appConf.Logger = &e2logrus.Config{
			Output:   "stdout",
			LogLevel: "debug",
		}
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

	logrus.Trace("logrus trace level active")
	logrus.Debug("logrus debug level active")
	logrus.Info("logrus info level active")
	logrus.Warnf("logrus warn level active")
	logrus.Error("logrus error level active")
}

func unmarshalAppConfig(v *viper.Viper) {
	if err := v.Unmarshal(&appConf); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if len(v.GetStringMap("general")) > 0 {
		appConf.General.PutAll(v.GetStringMap("general"))
	}
}
