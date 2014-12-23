package Assange

import (
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

func InitDb(config Configuration) *gorp.DbMap {
	var log = GetLogger("DB", DEBUG)
	source := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", config.Db_user, config.Db_password, config.Db_host, config.Db_database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		fmt.Println(err)
	}
	log.Info("Connect to database server:%s", config.Db_host)
	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
}

func InitTables(dbmap *gorp.DbMap) {
	dbmap.AddTable(TxInItem{}).SetKeys(true, "Id")
	dbmap.AddTable(TxOutItem{}).SetKeys(true, "Id")
	dbmap.AddTable(BlkHdrItem{}).SetKeys(true, "Id")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		fmt.Println(err)
	}
}
