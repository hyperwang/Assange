package main

import "haobtc/bitcoinrpc"
import "github.com/go-martini/martini"

var id int32 = 0
var server string
var m *martini.Martini

func getId() int32 {
	id += 1
	return id
}

func main() {
	server = "http://bitcoinrpc:Fe1H6fPTDtTXBKnUFdGyPE7w1CqsbScVmXoxHLbLTvDr@127.0.0.1:8332"
	bitcoinrpc.SearchRawTxs(server, getId(), "1JxMsgRGdKg3GQgpBabJWTUmyEDSPXdY1U", 0, 100)
	m = martini.New()
	r := martini.NewRouter()

	r.Get(`/queryapi/v1/unspent/:address`, GetUnspent)
	r.Post(`/queryapi/v1/unspent/:address`, GetUnspent)
	r.Get(`/queryapi/v1/tx/details/:txid`, GetDetails)
	r.Post(`/queryapi/v1/tx/details/:txid`, GetDetails)
	r.Get(`/queryapi/v1/tx/list/:address`, GetList)
	r.Post(`/queryapi/v1/tx/list/:address`, GetList)

	m.Action(r.Handle)
	m.Run()
}

func GetUnspent(params martini.Params) string {
	return "Unspent" + params["address"]
}

func GetDetails(params martini.Params) string {
	return "Details"
}

func GetList(params martini.Params) string {
	return bitcoinrpc.SearchRawTxs(server, getId(), params["address"], 0, 100)
}
