package cfg

import (
	"github.com/spf13/viper"
	"log"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/snoopd/")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Unable to read config:", err)
	}
}

var (
	GetBool = viper.GetBool
	GetString = viper.GetString
	GetInt = viper.GetInt
)
