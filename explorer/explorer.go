package explorer

import (
	. "Assange/config"
	. "Assange/logging"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
)

var ExplorerServer *martini.Martini
var dbmap *gorp.DbMap
var log = GetLogger("Explorer", DEBUG)

func InitExplorerServer(config Configuration) {
	source := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=True", config.Explorer_user, config.Explorer_password, config.Db_host, config.Db_database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		log.Error(err.Error())
	}
	log.Debug("Init explorer.")
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

	ExplorerServer = martini.New()

	r := martini.NewRouter()

	r.Get(`/api/v1/block/:hashid`, ApiBlockV1)
	r.Get(`/api/v1/tx/:hashid`, ApiTxV1)
	r.Get(`/api/v1/address/:addr`, ApiAddressV1)
	r.Get(`/api/v1/balance/:addr`, ApiBalanceV1)

	ExplorerServer.Action(r.Handle)
	ExplorerServer.RunOnAddr(":8000")
}
