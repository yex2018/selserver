package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yex2018/selserver/conf"
)

var SqlDB *sql.DB

func init() {
	var err error
	SqlDB, err = sql.Open(conf.Config.Sqlname, conf.Config.Mysql)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = SqlDB.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}
}
