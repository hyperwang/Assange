package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

var id int32 = 0

func getId() int32 {
    id += 1
    return id 
}

func main() {
    var server string
    server = "http://bitcoinrpc:Fe2H6fPTDtTXBKnUFdGyPE7w1CqsbScVmXoxHLbLTvDr@127.0.0.1:8332"
    BitcoinRPC(server,"getinfo",[]interface{}{})
    BitcoinRPC(server,"getblockhash",[]interface{}{1})
}

func BitcoinRPC(server string, method string, params []interface{}){
    data, err := json.Marshal(map[string]interface{}{
        "method":method,
        "id":getId(),
        "params":params,
    }) 
    if err != nil {
        log.Fatalf("Marshal: %v", err)
    }
    resp, err := http.Post(server,"application/json",strings.NewReader(string(data)))
    if err != nil {
        log.Fatalf("Post: %v", err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("ReadAll: %v", err)
    }
    result := make(map[string]interface{})
    err = json.Unmarshal(body, &result)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }
    log.Println(result)
}
