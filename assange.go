package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	. "Assange/explorer"
	. "Assange/logging"
	. "Assange/util"
	. "Assange/zmq"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"time"
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
var log = GetLogger("Main", WARNING)

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
		buildTx(dbmap)
	}
	//HandleZmq()
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

func buildTx(dbmap *gorp.DbMap) {
	for {
		var block_start_time time.Time
		var start_time time.Time

		block_start_time = time.Now()

		start_time = time.Now()
		trans, _ := dbmap.Begin()
		log.Warning("Checkpoint dbmap.Begin duration:%f.", time.Since(start_time).Seconds())

		//Get an unextracted block.
		start_time = time.Now()
		block, err := GetOneUnextractedBlock(trans)
		if err != nil {
			break
		}
		log.Warning("Checkpoint GetOneUnextractedBlock duration:%f.", time.Since(start_time).Seconds())

		//Make new Tx from rpc result, including spenditems for each Tx.
		for i := 0; i < len(block.Transactions); i += 32 {
			start_time = time.Now()
			bHash := block.Transactions[i : i+32]
			hash := hex.EncodeToString(ReverseBytes(bHash))
			rpcResult := RpcGetrawtransaction(hash)
			log.Warning("Checkpoint  RpcGetrawtransaction duration:%f.", time.Since(start_time).Seconds())

			start_time = time.Now()
			result, ok := rpcResult["result"].(string)
			tx := new(ModelTx)
			if ok {
				tx.ReceivedTime = block.Time
				if i == 0 {
					tx.IsCoinbase = true
				} else {
					tx.IsCoinbase = false
				}
				tx.Confirmed = true
				NewTxFromString(result, tx)
				tx = InsertTxIntoDB(trans, tx)
			} else {
				log.Error("Type assert error. tx.Hash:%s.", tx.Hash)
			}
			log.Warning("Checkpoint InsertTxIntoDB duration:%f.", time.Since(start_time).Seconds())

			//Maintain the relationship between block and tx
			start_time = time.Now()
			InsertRelationBlockTxIntoDB(trans, block, tx)
			log.Warning("Checkpoint03 InsertRelationBlockTxIntoDB duration:%f.", time.Since(start_time).Seconds())

			start_time = time.Now()
			for _, txout := range tx.Txouts {
				InsertTxoutIntoDb(trans, txout)
			}
			log.Warning("Checkpoint04 InsertTxoutIntoDb duration:%f.", time.Since(start_time).Seconds())

			start_time = time.Now()
			for _, txin := range tx.Txins {
				InsertTxinIntoDb(trans, txin)
			}
			log.Warning("Checkpoint05 InsertTxinIntoDb duration:%f.", time.Since(start_time).Seconds())

			tx.Extracted = true
			trans.Update(tx)
		}

		block.Extracted = true
		start_time = time.Now()
		trans.Update(block)
		log.Warning("Checkpoint06 trans.Update duration:%f.", time.Since(start_time).Seconds())

		start_time = time.Now()
		trans.Commit()
		log.Warning("Checkpoint07 trans.Commit duration:%f.", time.Since(start_time).Seconds())

		log.Warning("Block duration:%f", time.Since(block_start_time).Seconds())
	}
}
