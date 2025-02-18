package e2http

type Config struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
	BaseUrl string `mapstructure:"base_url"`
}
