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
var log = GetLogger("Main", DEBUG)

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
	err := InitTables(dbmap)
	if err != nil {
		log.Error(err.Error())
	}
	InitZmq(dbmap)
	go InitExplorerServer(Config)
	if reindexFlag {
		//buildBlockAndTxFromRpc(dbmap)
		buildBlock(dbmap)
	}
	HandleZmq()
}

func buildBlockAndTxFromRpc(dbmap *gorp.DbMap) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var hashFromIdx string
	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	//bcHeight = 50000
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
		InsertBlockIntoDb(trans, block)

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
				for _, txout := range tx.Txouts {
					InsertTxoutIntoDb(trans, txout)
				}
				for _, txin := range tx.Txins {
					InsertTxinIntoDb(trans, txin)
				}
			}
		}
		trans.Commit()
	}
}

func buildBlock(dbmap *gorp.DbMap) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var hashFromIdx string
	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	//bcHeight = 50000
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
		InsertBlockOnlyIntoDb(trans, block)

		trans.Commit()
		if dbHeight == bcHeight {
			bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
		}
	}
}
