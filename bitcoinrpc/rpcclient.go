package bitcoinrpc

import (
	"Assange/config"
	"Assange/logging"
	//"encoding/hex"
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"net/http"
	"strings"
)

var log = logging.GetLogger("RPC", logging.DEBUG)
var server string
var request_id int32

func InitRpcClient(config config.Configuration) {
	server = fmt.Sprintf("http://%s:%s@%s:%d",
		config.Rpc_user,
		config.Rpc_password,
		config.Rpc_host,
		config.Rpc_port)
	request_id = 0
}

func getRequestId() int32 {
	request_id += 1
	return request_id
}

func BitcoinRPC(method string, params []interface{}) map[string]interface{} {
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     getRequestId(),
		"params": params,
	})
	if err != nil {
		log.Error(err.Error())
	}
	resp, err := http.Post(server, "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Error(err.Error())
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	var x interface{}
	if err := d.Decode(&x); err != nil {
		log.Error(err.Error())
	}
	//fmt.Printf("decode to %#v\n", x)
	//result, err := json.Marshal(x)
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//fmt.Println(result)
	return x.(map[string]interface{})
}

func RpcGetinfo() map[string]interface{} {
	return BitcoinRPC("getinfo", []interface{}{})
}

func RpcGetblockhash(index int64) map[string]interface{} {
	return BitcoinRPC("getblockhash", []interface{}{index})
}

func RpcGetblock(hash string) map[string]interface{} {
	return BitcoinRPC("getblock", []interface{}{hash})
}

func RpcGetblockcount() map[string]interface{} {
	return BitcoinRPC("getblockcount", []interface{}{})
}
