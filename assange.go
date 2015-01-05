package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	. "Assange/explorer"
	. "Assange/logging"
	//. "Assange/util"
	. "Assange/zmq"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
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
	bcHeight = 1
	dbHeight, _ = GetMaxBlockHeightFromDB(dbmap)
	for dbHeight < bcHeight {
		dbHeight++

		//Get block info by height
		rpcResult = RpcGetblockhash(dbHeight)
		hashFromIdx = rpcResult["result"].(string)
		rpcResult = RpcGetblock(hashFromIdx)
		result := rpcResult["result"].(map[string]interface{})

		//Make new ModelBlock from rpc result
		block, _ := NewBlockFromMap(result)

		//Get raw transactions from rpc, parse to btcwire.MsgTx
		var msgtxs []*btcwire.MsgTx
		for _, tx := range block.Txs {
			rpcResult = RpcGetrawtransaction(tx.Hash)
			result, ok := rpcResult["result"].(string)
			if ok {
				bytesRawtx, _ := hex.DecodeString(result)
				tx, err := btcutil.NewTxFromBytes(bytesRawtx)
				if err != nil {
					log.Error(err.Error())
				}
				msgtx := tx.MsgTx()
				msgtxs = append(msgtxs, msgtx)
			}
		}

		trans, _ := dbmap.Begin()
		InsertBlockIntoDB(trans, block)
		if block.Height != 0 {
			for idx, tx := range block.Txs {
				sItems := NewSpendItemsFromMsg(msgtxs[idx], tx)
				InsertSpendItemsIntoDB(trans, sItems)
			}
		}
		trans.Commit()
	}
}
