package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	. "Assange/explorer"
	. "Assange/logging"
	//. "Assange/util"
	. "Assange/zmq"
	//"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	//"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	. "strconv"
	//"time"
)

var _ = ParseInt
var _ = fmt.Printf
var _ = json.Unmarshal
var Config config.Configuration
var log = GetLogger("Main", INFO)

var reindexFlag bool

func init() {
	const (
		defaultReindex = false
		usage          = "Regenerate database by bitcoind RPC."
	)
	flag.BoolVar(&reindexFlag, "reindex", defaultReindex, usage)
}

func main() {
	flag.Parse()
	Config, _ = config.InitConfiguration("config.json")
	InitRpcClient(Config)
	dbmap, _ := InitDb(Config)
	InitZmq()
	go HandleZmq()
	go InitExplorerServer(Config)
	if reindexFlag {
		err := InitTables(dbmap)
		if err != nil {
			log.Error(err.Error())
		}
		buildBlockAndTxFromRpc(dbmap)
	}
}

func buildBlockAndTxFromRpc(dbmap *gorp.DbMap) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var hashFromIdx string
	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	bcHeight = 50
	dbHeight, _ = GetMaxBlockHeightFromDB(dbmap)
	for dbHeight < bcHeight {
		dbHeight++
		trans, _ := dbmap.Begin()

		//Get block info by height
		rpcResult = RpcGetblockhash(dbHeight)
		hashFromIdx = rpcResult["result"].(string)
		rpcResult = RpcGetblock(hashFromIdx)
		result := rpcResult["result"].(map[string]interface{})

		//Make new ModelBlock from rpc result
		block, _ := NewBlockFromMap(result)
		//Insert into DB, and tx's id will be updated
		InsertBlockIntoDB(trans, block)

		//Make new Tx from rpc result, including spenditems for each Tx.
		for _, tx := range block.Txs {
			rpcResult = RpcGetrawtransaction(tx.Hash)
			result, ok := rpcResult["result"].(string)
			if ok {
				NewTxFromString(result, tx)
			} else {
				log.Error("Type assert error. tx.Hash:%s.", tx.Hash)
			}
		}
		//Insert spenditem into db.
		if block.Height != 0 {
			for _, tx := range block.Txs {
				InsertSpendItemsIntoDB(trans, tx.SItems)
			}
		}
		trans.Commit()
	}
}
