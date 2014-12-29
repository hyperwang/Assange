package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	. "Assange/logging"
	. "Assange/util"
	"encoding/hex"
	"encoding/json"
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

func main() {
	Config, _ = config.InitConfiguration("config.json")
	dbmap, _ := InitDb(Config)
	err := InitTables(dbmap)
	if err != nil {
		log.Error(err.Error())
	}
	InitRpcClient(Config)
	buildBlockAndTxFromRpc(dbmap)
}

func buildBlockAndTxFromRpc(dbmap *gorp.DbMap) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var block *ModelBlock
	var hashFromIdx string
	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	//bcHeight = 170
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
				fmt.Println(result)
				fmt.Println(bytesRawtx)
				tx, err := btcutil.NewTxFromBytes(bytesRawtx)
				if err != nil {
					fmt.Println(err)
				}
				msgtx := tx.MsgTx()
				msgtxs = append(msgtxs, msgtx)
				fmt.Println("Append to msgtxs.")
			}
		}

		trans, _ := dbmap.Begin()
		NewBlockIntoDB(trans, block, txs)
		fmt.Println(len(txs), len(msgtxs))
		if block.Height != 0 {
			for idx, tx := range txs {
				NewSpendItemIntoDB(trans, msgtxs[idx], tx)
			}
		}
		trans.Commit()
	}
}
