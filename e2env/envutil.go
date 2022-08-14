package e2env

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 從環境變量或命令行參數獲取參數值
// 需要在調用方執行 flag.Parse()
// 命令行參數的 "-" 對應環境變量的 "_"
// 用法:
// var p string
// e2env.EnvStringVar(&p,"param-name","default value","usage .....")
// flag.Parse()
// 上述代碼將可從命令行參數 --param-name=xxxxx 或環境變量 PARAM_NAME=xxxxx 取值

func convertEnvKey(key string) string {
	return strings.ToUpper(strings.Replace(key, "-", "_", -1))
}

// EnvStringVar 从命令行参数或环环境变量取参数,优先取环境变量值
func EnvStringVar(p *string, key string, defaultVal string, usage string) {
	if v, ok := os.LookupEnv(convertEnvKey(key)); ok && v != "" {
		*p = v
		return
	}
	if flag.Lookup(key) == nil {
		flag.StringVar(p, key, defaultVal, fmt.Sprintf("%s=%s ,%s", convertEnvKey(key), defaultVal, usage))
	}

	*p = flag.Lookup(key).Value.(flag.Getter).Get().(string)
}

// EnvBoolVar 从命令行参数或环环境变量取参数,优先取环境变量值
func EnvBoolVar(p *bool, key string, defaultVal bool, usage string) {
	if v, ok := os.LookupEnv(convertEnvKey(key)); ok && v != "" {
		if ev, err := strconv.ParseBool(v); err == nil {
			*p = ev
			return
		}
	}

	if flag.Lookup(key) == nil {
		flag.BoolVar(p, key, defaultVal, fmt.Sprintf("%s=%v ,%s", convertEnvKey(key), defaultVal, usage))
	}

	*p = flag.Lookup(key).Value.(flag.Getter).Get().(bool)
}

// EnvIntVar 从命令行参数或环环境变量取参数,优先取环境变量值
func EnvIntVar(p *int, key string, defaultVal int, usage string) {

	if v, ok := os.LookupEnv(convertEnvKey(key)); ok && v != "" {
		if ev, err := strconv.Atoi(v); err == nil {
			*p = ev
			return
		}
	}

	if flag.Lookup(key) == nil {
		flag.IntVar(p, key, defaultVal, fmt.Sprintf("%s=%v ,%s", convertEnvKey(key), defaultVal, usage))
	}

	*p = flag.Lookup(key).Value.(flag.Getter).Get().(int)
}
