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
	"sync"
	"time"
	//"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
	"github.com/coopernurse/gorp"
	. "strconv"
	//"time"
)

var _ = time.Now
var _ = ParseInt
var _ = fmt.Printf
var _ = json.Unmarshal
var Config config.Configuration
var log = GetLogger("Main", DEBUG)

var buildblockFlag bool
var checkblockFlag bool

func init() {
	const (
		buildblockDefault = false
		buildblockUsage   = "Regenerate database by bitcoind RPC."

		checkblockDefault = false
		checkblockUsage   = "Check the relationship between block and tx."
	)
	flag.BoolVar(&buildblockFlag, "buildblock", buildblockDefault, buildblockUsage)
	flag.BoolVar(&checkblockFlag, "checkblock", checkblockDefault, checkblockUsage)
}

func main() {
	flag.Parse()
	var wait sync.WaitGroup
	wait.Add(1)
	Config, _ = config.InitConfiguration("config.json")
	InitRpcClient(Config)
	dbmap, _ := InitDb(Config)
	InitTables(dbmap)
	InitZmq(dbmap)
	go InitExplorerServer(Config)
	if buildblockFlag {
		buildBlock(dbmap, 50000)
		buildTxFromBlock(dbmap)
		extractTx(dbmap)
		extractTxout(dbmap)
		extractTxin(dbmap)
	}
	if checkblockFlag {
		checkBlock(dbmap)
	}
	wait.Wait()
	//HandleZmq()
}

func buildBlock(dbmap *gorp.DbMap, height int64) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var hashFromIdx string
	if height == 0 {
		bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	} else {
		bcHeight = height
	}
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
		block := new(ModelBlock)
		block.NewFromMap(result)
		//Insert into DB, and tx's id will be updated
		//InsertBlockOnlyIntoDb(trans, block)
		block.InsertIntoDb(trans)

		trans.Commit()
	}
}

func buildTxFromBlock(dbmap *gorp.DbMap) {
	for {
		trans, _ := dbmap.Begin()

		//Get an unextracted block.
		block := new(ModelBlock)
		err := block.NewFromUnextracted(trans)
		if err != nil {
			break
		}

		for idx, txnHash := range RpcGetblockTxns(block.Hash) {
			resp := RpcGetrawtransaction(txnHash)
			result, ok := resp["result"].(string)
			if !ok {
				log.Error("Type assert error. tx.Hash:%s.", txnHash)
				continue
			}
			tx := new(ModelTx)
			tx.NewFromString(result)
			tx.ReceivedTime = block.Time
			if idx == 0 {
				tx.IsCoinbase = true
			} else {
				tx.IsCoinbase = false
			}
			tx.Confirmed = true
			err := tx.InsertIntoDb(trans)
			if err != nil {
				log.Error(err.Error())
			}

			//Maintain the relationship between block and tx
			r := new(RelationBlockTx)
			r.InsertIntoDb(trans, block, tx)
			block.Extracted = true
			trans.Update(block)
		}
		trans.Commit()
	}
}

func extractTx(dbmap *gorp.DbMap) {
	for {
		trans, _ := dbmap.Begin()
		tx := new(ModelTx)
		err := tx.NewFromUnextracted(trans)
		if err != nil {
			break
		}
		resp := RpcGetrawtransaction(tx.Hash)
		result, ok := resp["result"].(string)
		tx1 := new(ModelTx)
		if ok {
			tx.UpdateInOutFromString(result)
		} else {
			log.Error("Type assert error. tx.Hash:%s.", tx1.Hash)
			continue
		}

		ins := new(ModelTxinSet)
		ins.NewFromTx(tx)
		if err := ins.InsertIntoDb(trans); err != nil {
			log.Error(err.Error())
		}

		outs := new(ModelTxoutSet)
		outs.NewFromTx(tx)
		if err := outs.InsertIntoDb(trans); err != nil {
			log.Error(err.Error())
		}

		//Update tx to extracted.
		tx.Extracted = true
		trans.Update(tx)
		trans.Commit()
		log.Info("Tx extraced. Id:%d Hash:%s.", tx.Id, tx.Hash)
	}
}

func extractTxout(dbmap *gorp.DbMap) {
	for {
		trans, _ := dbmap.Begin()
		txout := new(ModelTxout)
		err := txout.NewFromUnextracted(trans)
		if err != nil {
			break
		}

		class, addresses, reqSig, _ := btcscript.ExtractPkScriptAddrs(txout.OutScript, &btcnet.MainNetParams)
		txout.Type = class
		txout.ReqSig = reqSig

		for _, address := range addresses {
			mAddress := new(ModelAddress)
			mAddress.UpdateFromDbByAddress(trans, address.EncodeAddress())
			trans.Update(mAddress)
			r := new(RelationTxoutAddress)
			r.InsertIntoDb(trans, txout, mAddress)
			if txout.Type >= btcscript.PubKeyTy && txout.Type <= btcscript.ScriptHashTy {
				mAddress.Balance += txout.Value
				trans.Update(mAddress)
			}
		}
		txout.Extracted = true
		trans.Update(txout)
		trans.Commit()
	}
}

func extractTxin(dbmap *gorp.DbMap) {
	for {
		trans, _ := dbmap.Begin()
		txin := new(ModelTxin)
		err := txin.NewFromUncalculated(trans)
		if err != nil {
			break
		}
		if txin.IsCoinbase {
			log.Info("Txin is from coinbase,skip.")
			txin.Calculated = true
			trans.Update(txin)
			trans.Commit()
			continue
		}
		var txout = new(ModelTxout)
		query := fmt.Sprintf("select * from txout where OutTxHash=\"%s\" and OutIndex=%d", txin.PrevOutHash, txin.PrevOutIndex)
		err = trans.SelectOne(txout, query)
		if err != nil {
			log.Error(err.Error())
		}
		if txout.Id == 0 {
			log.Info("No matched txout found.")
			txin.Calculated = true
			trans.Update(txin)
			trans.Commit()
			continue
		} else {
			log.Info("Txout matched, Id:%d, Hash:%s, Index:%d.", txout.Id, txout.OutTxHash, txout.OutIndex)
			txin.Calculated = true
			trans.Update(txin)
			if txout.Type >= btcscript.PubKeyTy && txout.Type <= btcscript.ScriptHashTy {
				address := new(ModelAddress)
				err := trans.SelectOne(&address, "select * from address where Id in (select AddressId from txoutaddress where TxoutId=?)", txout.Id)
				if err != nil {
					log.Error(err.Error())
				}
				address.Balance -= txout.Value
				trans.Update(address)
			}
			txout.Spent = true
			txout.RefTxinId = txin.Id
			trans.Update(txout)
			trans.Commit()
		}
	}
}

func checkBlock(dbmap *gorp.DbMap) {
	blockId, _ := GetMaxBlockIdFromDB(dbmap)
	log.Error("Hello world,blockId:%d", blockId)
	var i int64
	for i = 1; i < blockId; i++ {
		var block []*ModelBlock
		_, err := dbmap.Select(&block, "select * from block where Id=?", i)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if len(block) == 0 {
			continue
		}
		log.Debug(block[0].Hash)
	}
	return
}

func checkTx(dbmap *gorp.DbMap) {
	txId, _ := GetMaxTxIdFromDB(dbmap)
	var i int64
	for i = 1; i < txId; i++ {
		txHash, err := dbmap.SelectStr("select Hash from tx where Id=?", i)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		RpcGetrawtransaction(txHash)
	}
}
