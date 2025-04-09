package main

import (
	"cache/service"
	"fmt"
)

func main() {
	s := service.NewServer("127.0.0.1", 8999)
	fmt.Println("========================================")
	fmt.Println("  ______   _____           _          \n |___  /  / ____|         | |         \n    / /  | |     __ _  ___| |__   ___ \n   / /   | |    / _` |/ __| '_ \\ / _ \\\n  / /__  | |___| (_| | (__| | | |  __/\n /_____|  \\_____\\__,_|\\___|_| |_|\\___|")
	fmt.Println("========================================")
	fmt.Println("version:v0.1 beta")
	fmt.Println("server run at ", "127.0.0.1:8999")
	s.Run()
}
