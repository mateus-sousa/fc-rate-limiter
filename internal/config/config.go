package config

import (
	"github.com/spf13/viper"
)

type Conf struct {
	ReqPerSecondsToken int `mapstructure:"REQ_PER_SEC_TOKEN"`
	ReqPerSecondsIP    int `mapstructure:"REQ_PER_SEC_IP"`
}

func LoadConfig(path string) (*Conf, error) {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	var cfg *Conf
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg, nil
}
