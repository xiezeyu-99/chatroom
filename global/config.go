package global

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var SensitiveWords []string

func initConfig() {
	viper.SetConfigName("chatroom")
	viper.AddConfigPath(RootDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	SensitiveWords = viper.GetStringSlice("sensitive")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
		SensitiveWords = viper.GetStringSlice("sensitive")
		log.Println("重新加载配置")
	})
}
