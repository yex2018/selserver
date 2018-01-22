package main

import (
	"github.com/yex2018/selserver/router"

	"github.com/yex2018/selserver/conf"
	db "github.com/yex2018/selserver/database"
)

func main() {
	defer db.SqlDB.Close()
	router := router.InitRouter()
	router.Run(conf.Config.Port)
}
