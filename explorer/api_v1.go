package explorer

import (
	. "Assange/blockdata"
	//. "Assange/util"
	"encoding/hex"
	"encoding/json"
	"github.com/conformal/btcscript"
	"github.com/go-martini/martini"
	"net/http"
)

type BlockV1 struct {
	Hash       string   `json:"hash"`
	Height     int64    `json:"height"`
	Ver        int32    `json:"version"`
	Time       int64    `json:"time"`
	PrevHash   string   `json:"prev_hash"`
	NextHash   string   `json:"next_hash"`
	Nonce      uint32   `json:"nonce"`
	Bits       uint32   `json:"bits"`
	MerkleRoot string   `json:"merkle_root"`
	Txn        []string `json:"transaction"`
}

type BalanceV1 struct {
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

type TxV1 struct {
	Hash     string     `json:"hash"`
	Ver      int32      `json:"version"`
	LockTime uint32     `json:"lock_time"`
	Txin     []*TxinV1  `json:"input"`
	Txout    []*TxoutV1 `json:"output"`
	Block    []string   `json:"block"`
}

type TxinV1 struct {
	Sequence uint32
	Script   string
}

type TxoutV1 struct {
	Value  int64
	Script string
	Index  int64
	Type   btcscript.ScriptClass
	Spent  bool
}

func ApiBlockV1(params martini.Params) (int, string) {
	return http.StatusOK, GetBlockV1(params["hashid"])
}

func ApiTxV1(params martini.Params) (int, string) {
	return http.StatusOK, GetTxV1(params["hashid"])
}

func ApiAddressV1(params martini.Params) (int, string) {
	return http.StatusOK, params["addr"]
}

func ApiBalanceV1(params martini.Params) (int, string) {
	return http.StatusOK, GetBalanceV1(params["addr"])
}

func GetBlockV1(hash string) string {
	var block = new(BlockV1)
	var blockBuff []*ModelBlock

	oldLen := len(blockBuff)
	_, err := dbmap.Select(&blockBuff, "select * from block where Hash=?", hash)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	if oldLen == len(blockBuff) {
		log.Error("Block not found. Hash:%s.", hash)
		return "Error"
	}
	block.Hash = blockBuff[0].Hash
	block.Height = blockBuff[0].Height
	block.Ver = blockBuff[0].Ver
	block.Time = blockBuff[0].Time.Unix()
	block.PrevHash = blockBuff[0].PrevHash
	block.NextHash = blockBuff[0].NextHash
	block.Nonce = blockBuff[0].Nonce
	block.Bits = blockBuff[0].Bits
	block.MerkleRoot = blockBuff[0].MerkleRoot

	var txHash []string
	_, err = dbmap.Select(&txHash, "select Hash from tx where Id in (select TxId from blocktx where BlockId=?)", blockBuff[0].Id)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	block.Txn = txHash

	jsonBytes, err := json.MarshalIndent(block, "", "    ")
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	return string(jsonBytes)
}

func GetBalanceV1(addr string) string {
	var balanceMap = new(BalanceV1)
	balance, err := dbmap.SelectInt("select Balance from address where Address=?", addr)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	balanceMap.Address = addr
	balanceMap.Balance = balance
	jsonBytes, err := json.MarshalIndent(balanceMap, "", "    ")
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	return string(jsonBytes)
}

func GetTxV1(hashid string) string {
	var txMap = new(TxV1)
	var tx = new(ModelTx)

	err := dbmap.SelectOne(&tx, "select * from tx where Hash=?", hashid)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	txMap.Hash = tx.Hash
	txMap.Ver = tx.Ver
	txMap.LockTime = tx.LockTime

	var inBuff []*ModelTxin
	_, err = dbmap.Select(&inBuff, "select * from txin where InTxHash=?", hashid)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	for _, in := range inBuff {
		txinMap := new(TxinV1)
		txinMap.Sequence = in.Sequence
		txinMap.Script = hex.EncodeToString(in.InScript)
		txMap.Txin = append(txMap.Txin, txinMap)
	}

	var outBuff []*ModelTxout
	_, err = dbmap.Select(&outBuff, "select * from txout where OutTxHash=?", hashid)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	for _, out := range outBuff {
		txoutMap := new(TxoutV1)
		txoutMap.Type = out.Type
		txoutMap.Spent = out.Spent
		txoutMap.Value = out.Value
		txoutMap.Script = hex.EncodeToString(out.OutScript)
		txoutMap.Index = out.OutIndex
		txMap.Txout = append(txMap.Txout, txoutMap)
	}

	var mBlock = new(ModelBlock)
	err = dbmap.SelectOne(mBlock, "select Hash from block where Id in (select BlockId from blocktx where TxId=? )", tx.Id)
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	txMap.Block = append(txMap.Block, mBlock.Hash)
	jsonBytes, err := json.MarshalIndent(txMap, "", "    ")
	if err != nil {
		log.Error(err.Error())
		return "Error"
	}
	return string(jsonBytes)
}
