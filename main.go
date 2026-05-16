package main

import (
	"LanshanSummerProject/app/api/configs"
	"LanshanSummerProject/app/api/router"
	"fmt"
)

func main() {
	configs.InitLogger()
	defer func() {
		if err := configs.Logger.Sync(); err != nil {
			fmt.Println(err)
		}
	}()
	configs.InitDB()
	router.Start()
}
