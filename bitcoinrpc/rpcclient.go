package bitcoinrpc 

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

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
