package Assange

import (
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
)

func InitDb(config Configuration) *gorp.DbMap {
	source := fmt.Sprintf("tcp:%s:3306*%s/%s/%s", config.Db_host, config.Db_database, config.Db_user, config.Db_password)
	db, err := sql.Open("mysql", source)
	if err != nil {
		fmt.Println(err)
	}
	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
}
