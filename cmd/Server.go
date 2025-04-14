package main

import (
	"cache/service"
	"fmt"
	"github.com/spf13/viper"
	"strconv"
)

func main() {
	//首先加载配置文件
	LoadConfig()
	s := service.NewServer(
		viper.GetString("ip"),
		viper.GetInt("port"),
		viper.GetBool("persistence"),
		viper.GetInt("persistence-time"),
	)
	fmt.Println("========================================")
	fmt.Println("  ______   _____           _          \n |___  /  / ____|         | |         \n    / /  | |     __ _  ___| |__   ___ \n   / /   | |    / _` |/ __| '_ \\ / _ \\\n  / /__  | |___| (_| | (__| | | |  __/\n /_____|  \\_____\\__,_|\\___|_| |_|\\___|")
	fmt.Println("========================================")
	fmt.Println("version : v0.2 beta")
	fmt.Println("server listen at ", viper.GetString("ip"), ":", strconv.Itoa(viper.GetInt("port")))
	fmt.Println("persistence : ", viper.GetBool("persistence"))
	s.Run()
}

func LoadConfig() {
	viper.SetConfigName("config.yml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
