package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	. "Assange/zmq"
	//. "Assange/explorer"
	. "Assange/logging"
	. "Assange/util"
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
	InitZmq()
	HandleZmq()
	//InitExplorerServer(Config)
	if reindexFlag {
		dbmap, _ := InitDb(Config)
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
	var block *ModelBlock
	var hashFromIdx string
	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	bcHeight = 50000
	dbHeight, _ = GetMaxBlockHeightFromDB(dbmap)
	for dbHeight < bcHeight {
		dbHeight++

		//Get block info by height
		rpcResult = RpcGetblockhash(dbHeight)
		hashFromIdx = rpcResult["result"].(string)
		rpcResult = RpcGetblock(hashFromIdx)
		result := rpcResult["result"].(map[string]interface{})

		//New ModelBlock and ModelTx
		block = new(ModelBlock)
		txs, _ := block.NewBlock(result)

		//Get raw transactions from rpc, parse to btcwire.MsgTx
		var msgtxs []*btcwire.MsgTx
		for _, tx := range txs {
			rpcResult = RpcGetrawtransaction(hex.EncodeToString(ReverseBytes(tx.Hash)))
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
		NewBlockIntoDB(trans, block, txs)
		if block.Height != 0 {
			for idx, tx := range txs {
				NewSpendItemIntoDB(trans, msgtxs[idx], tx)
			}
		}
		trans.Commit()
	}
}
