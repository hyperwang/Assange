package explorer

import (
	. "Assange/config"
	. "Assange/logging"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
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

	r.Get(`/api/v1/block/:hashid`, ApiBlock)
	r.Get(`/api/v1/tx/:hashid`, ApiTx)
	r.Get(`/api/v1/address/:addr`, ApiAddress)
	r.Get(`/api/v1/balance/:addr`, ApiBalance)

	ExplorerServer.Action(r.Handle)
	ExplorerServer.RunOnAddr(":8000")
}

func ApiBlock(params martini.Params) (int, string) {
	return http.StatusOK, params["hashid"]
}

func ApiTx(params martini.Params) (int, string) {
	return http.StatusOK, params["hashid"]
}

func ApiAddress(params martini.Params) (int, string) {
	return http.StatusOK, params["addr"]
}

func ApiBalance(params martini.Params) (int, string) {
	return http.StatusOK, GetBalance(params["addr"])
}

func GetBalance(addr string) string {
	balance, err := dbmap.SelectInt("select Balance from balance where Address=?", addr)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	data := map[string]interface{}{
		"address": addr,
		"balance": balance,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	return string(jsonBytes)
}
