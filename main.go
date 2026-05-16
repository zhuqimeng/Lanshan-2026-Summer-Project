package main

import (
	"LanshanSummerProject/app/api/configs"
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
}
