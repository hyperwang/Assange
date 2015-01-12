package blockdata

import (
	//"Assange/config"
	//. "Assange/util"
	//"database/sql"
	//"encoding/hex"
	"encoding/json"
	//"github.com/conformal/btcutil"
	//"errors"
	//"github.com/conformal/btcscript"
	//"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	. "strconv"
	"time"
)

type ModelBlock struct {
	Id int64

	//Block info
	Height     int64
	Hash       string
	PrevHash   string
	NextHash   string
	MerkleRoot string
	Time       time.Time
	Ver        int32
	Nonce      uint32
	Bits       uint32

	//Transactions
	Txs []*ModelTx `db:"-"`

	//More flags to be added
	Extracted bool
}

func (block *ModelBlock) InsertIntoDb(trans *gorp.Transaction) {
	var prevBlock *ModelBlock
	if block.Height != 0 {
		err := trans.SelectOne(&prevBlock, "select * from block where Hash=?", block.PrevHash)
		if err != nil {
			log.Error(err.Error())
			return
		}
		prevBlock.NextHash = block.Hash
		trans.Update(prevBlock)
	} else {
		block.Extracted = true
	}
	err := trans.Insert(block)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("Insert new block, Id:%d, Height:%d.", block.Id, block.Height)
}

func (block *ModelBlock) NewFromUnextracted(trans *gorp.Transaction) error {
	err := trans.SelectOne(block, "select * from block where Extracted=0 limit 1")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Unextracted block found. Id:%d Hash:%s.", block.Id, block.Hash)
	return nil
}

func (block *ModelBlock) NewFromMap(resultMap map[string]interface{}) error {
	block.Height, _ = ParseInt(string(resultMap["height"].(json.Number)), 10, 64)
	block.Hash = resultMap["hash"].(string)
	if _, ok := resultMap["previousblockhash"]; ok {
		block.PrevHash = resultMap["previousblockhash"].(string)
	}
	if _, ok := resultMap["nextblockhash"]; ok {
		block.NextHash = resultMap["nextblockhash"].(string)
	}
	block.MerkleRoot = resultMap["merkleroot"].(string)
	timeUint64, _ := ParseInt(string(resultMap["time"].(json.Number)), 10, 64)
	block.Time = time.Unix(timeUint64, 0)
	verUint64, _ := ParseUint(string(resultMap["version"].(json.Number)), 10, 32)
	block.Ver = int32(verUint64)
	nonceUint64, _ := ParseUint(string(resultMap["nonce"].(json.Number)), 10, 32)
	block.Nonce = uint32(nonceUint64)
	bitsUint64, _ := ParseUint(resultMap["bits"].(string), 16, 32)
	block.Bits = uint32(bitsUint64)
	txsResult, _ := resultMap["tx"].([]interface{})
	for idx, txResult := range txsResult {
		tx := new(ModelTx)
		tx.Hash = txResult.(string)
		tx.ReceivedTime = block.Time
		tx.Confirmed = true
		tx.Extracted = false
		if idx == 0 {
			tx.IsCoinbase = true
		} else {
			tx.IsCoinbase = false
		}
		block.Txs = append(block.Txs, tx)
	}
	return nil
}

func (block *ModelBlock) NewFromRpcByHeight(height int64) {
}

func (block *ModelBlock) NewFromRpcByHas() {
}

func InitModelBlockTable(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(ModelBlock{}, "block").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index uidx_block_hash on block(Hash)")
	dbmap.Exec("create index idx_block_prevhash on block(PrevHash)")
	dbmap.Exec("create index idx_block_nexthash on block(NextHash)")
	dbmap.Exec("create index idx_block_height on block(Height)")
	dbmap.Exec("create index idx_block_extracted on block(Extracted)")
}
