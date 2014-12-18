package bitcoinrpc

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

func BitcoinRPC(server string, method string, id int32, params []interface{}) string {
    data, err := json.Marshal(map[string]interface{}{
        "method":method,
        "id":id,
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
    result := string(body)
    log.Println(result)
    /*result := make(map[string]interface{})
    err = json.Unmarshal(body, &result)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }*/
    return result
}
