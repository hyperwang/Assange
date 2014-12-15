package main

import  "haobtc/bitcoinrpc"

var id int32 = 0

func getId() int32 {
    id += 1
    return id 
}

func main() {
    var server string
    server = "http://bitcoinrpc:Fe2H6fPTDtTXBKnUFdGyPE7w1CqsbScVmXoxHLbLTvDr@127.0.0.1:8332"
    //bitcoinrpc.BitcoinRPC(server,"getinfo", getId(), []interface{}{})
    //bitcoinrpc.BitcoinRPC(server,"getblockhash", getId(), []interface{}{1})
    //bitcoinrpc.BitcoinRPC(server,"searchrawtransactions", getId(), []interface{}{"1JxMsgRGdKg3GQgpBabJWTUmyEDSPXdY1U",1})
    bitcoinrpc.SearchRawTxs(server, getId(), "1JxMsgRGdKg3GQgpBabJWTUmyEDSPXdY1U",0,100)
}

