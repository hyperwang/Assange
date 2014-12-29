package explorer

import (
	"github.com/go-martini/martini"
	"net/http"
)

var ExplorerServer *martini.Martini

func InitExplorerServer() {
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
	return http.StatusOK, params["addr"]
}
