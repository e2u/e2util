package e2http

import (
	"github.com/e2u/e2util/e2logrus"
)

type Config struct {
	Address      string           `mapstructure:"address"`
	Port         int              `mapstructure:"port"`
	BaseUrl      string           `mapstructure:"base_url"`
	LoggerConfig *e2logrus.Config `mapstructure:"logger"`
}

func (c *Config) GetLoggerFormat() string {
	if c.LoggerConfig == nil {
		return "text"
	}
	return c.LoggerConfig.Format
}
