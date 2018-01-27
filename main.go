package main

import (
	"github.com/yex2018/selserver/conf"
	"github.com/yex2018/selserver/database"
	"github.com/yex2018/selserver/router"
)

func main() {
	defer database.SqlDB.Close()
	router := router.InitRouter()
	router.Run(conf.Config.Port)
}
