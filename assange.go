package main

import (
	. "Assange/bitcoinrpc"
	. "Assange/blockdata"
	"Assange/config"
	"encoding/hex"
	"encoding/json"
	//"fmt"
	"github.com/coopernurse/gorp"
	. "strconv"
	"time"
)

var Config config.Configuration

func main() {
	Config, _ = config.InitConfiguration("config.json")
	dbmap, _ := InitDb(Config)
	InitTables(dbmap)
	InitRpcClient(Config)
	GetMaxBlockHeightFromDB(dbmap)
	//block := new(ModelBlock)
	//block.Height = uint32(1000)
	buildBlockAndTxFromRpc(dbmap)
}

func buildBlockAndTxFromRpc(dbmap *gorp.DbMap) {
	var bcHeight int64
	var dbHeight int64
	var rpcResult map[string]interface{}
	var block *ModelBlock
	var hashFromIdx string

	bcHeight, _ = ParseInt(string(RpcGetblockcount()["result"].(json.Number)), 10, 64)
	//bcHeight = 2
	dbHeight, _ = GetMaxBlockHeightFromDB(dbmap)
	for dbHeight < bcHeight {
		dbHeight++

		//Get block info by height
		rpcResult = RpcGetblockhash(dbHeight)
		hashFromIdx = rpcResult["result"].(string)
		rpcResult = RpcGetblock(hashFromIdx)
		result := rpcResult["result"].(map[string]interface{})

		//Parse rpc result to a new block
		block = new(ModelBlock)
		block.Height, _ = ParseInt(string(result["height"].(json.Number)), 10, 64)
		block.Hash, _ = hex.DecodeString(result["hash"].(string))
		if _, ok := result["previousblockhash"]; ok {
			block.PrevHash, _ = hex.DecodeString(result["previousblockhash"].(string))
		}
		if _, ok := result["nextblockhash"]; ok {
			block.NextHash, _ = hex.DecodeString(result["nextblockhash"].(string))
		}
		block.MerkleRoot, _ = hex.DecodeString(result["merkleroot"].(string))
		timeUint64, _ := ParseInt(string(result["time"].(json.Number)), 10, 64)
		block.Time = time.Unix(timeUint64, 0)
		verUint64, _ := ParseUint(string(result["version"].(json.Number)), 10, 32)
		block.Ver = uint32(verUint64)
		nonceUint64, _ := ParseUint(string(result["nonce"].(json.Number)), 10, 32)
		block.Nonce = uint32(nonceUint64)
		bitsUint64, _ := ParseUint(result["bits"].(string), 16, 32)
		block.Bits = uint32(bitsUint64)

		//Insert new block
		NewBlockIntoDB(dbmap, block, nil)
	}
}
