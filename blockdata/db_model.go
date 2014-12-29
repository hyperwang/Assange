package blockdata

import (
	"Assange/config"
	. "Assange/logging"
	. "Assange/util"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"github.com/conformal/btcutil"
	"errors"
	"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	. "strconv"
	"time"
)

type ModelBlock struct {
	Id int64

	//Block info
	Height     int64
	Hash       []byte
	PrevHash   []byte
	NextHash   []byte
	MerkleRoot []byte
	Time       time.Time
	Ver        uint32
	Nonce      uint32
	Bits       uint32

	//More flags to be added
	ConfirmFlag bool
}

type ModelTx struct {
	Id int64
	//BlockId int64

	//Transaction info
	Hash []byte

	//More flags to be added
}

type RelationBlockTx struct {
	Id int64

	//Including relation between block and tx
	BlockId int64
	TxId    int64

	//More flags to be added
}

type ModelAddressBalance struct {
	Id int64

	Address string
	Balance int64
}

type ModelSpendItem struct {
	Id int64

	//Transaction output info
	OutTxId   int64
	Type      int64
	OutScript []byte
	Addr      []byte
	Value     int64
	Index     int64

	//Transaction input info
	InTxId   int64
	InScript []byte

	//More flags to be added
}

type ModelTxin struct {
	Id   int64
	TxId int64

	//Transaction input info
	TxHash       []byte
	PrevTxHash   []byte
	PrevOutIndex int64

	//More flags to be added
}

var log = GetLogger("DB", DEBUG)

func InitDb(config config.Configuration) (*gorp.DbMap, error) {
	source := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=True", config.Db_user, config.Db_password, config.Db_host, config.Db_database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		return nil, err
	}
	log.Info("Connect to database server:%s", config.Db_host)
	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}, nil
}

func InitTables(dbmap *gorp.DbMap) error {
	dbmap.AddTableWithName(ModelBlock{}, "block").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelTx{}, "tx").SetKeys(true, "Id")
	dbmap.AddTableWithName(RelationBlockTx{}, "block_tx").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelSpendItem{}, "spend_item").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelTxin{}, "txin").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelAddressBalance{}, "balance").SetKeys(true, "Id")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}
	return nil
}

func GetMaxBlockHeightFromDB(dbmap *gorp.DbMap) (int64, error) {
	//var blkBuff []*ModleBlock
	var maxHeight int64
	maxHeight, _ = dbmap.SelectInt("select max(Height) as Height from block")
	if maxHeight == 0 {
		rowCount, _ := dbmap.SelectInt("select count(*) from block")
		if rowCount == 0 {
			maxHeight = -1
			log.Info("No record in block database. Set max height to -1.")
		}
	}
	log.Info("Max height in block database is %d.", maxHeight)
	return maxHeight, nil
}

func GetMaxTxIdFromDB(dbmap *gorp.DbMap) (int64, error) {
	var maxTxId int64
	maxTxId, _ = dbmap.SelectInt("select max(Id) as Id from tx")
	log.Info("Max id in transaction database is %d.", maxTxId)
	return maxTxId, nil
}

func NewBlockIntoDB(trans *gorp.Transaction, block *ModelBlock, tx []*ModelTx) error {
	//trans, _ := dbmap.Begin()
	block.ConfirmFlag = true
	trans.Insert(block)
	log.Info("Block Id:%d", block.Id)
	for idx, _ := range tx {
		//tx[idx].BlockId = block.Id
		err := trans.Insert(tx[idx])
		if err != nil {
			log.Error(err.Error())
		}
		blockTx := new(RelationBlockTx)
		blockTx.BlockId = block.Id
		blockTx.TxId = tx[idx].Id
		err = trans.Insert(blockTx)
		if err != nil {
			log.Error(err.Error())
		}
	}
	//trans.Commit()
	return nil
}

func (block *ModelBlock) NewBlock(resultMap map[string]interface{}) ([]*ModelTx, error) {
	block.Height, _ = ParseInt(string(resultMap["height"].(json.Number)), 10, 64)
	bytesBuff, _ := hex.DecodeString(resultMap["hash"].(string))
	block.Hash = ReverseBytes(bytesBuff)
	if _, ok := resultMap["previousblockhash"]; ok {

		bytesBuff, _ := hex.DecodeString(resultMap["previousblockhash"].(string))
		block.PrevHash = ReverseBytes(bytesBuff)
	}
	if _, ok := resultMap["nextblockhash"]; ok {
		bytesBuff, _ = hex.DecodeString(resultMap["nextblockhash"].(string))
		block.NextHash = ReverseBytes(bytesBuff)
	}
	bytesBuff, _ = hex.DecodeString(resultMap["merkleroot"].(string))
	block.MerkleRoot = ReverseBytes(bytesBuff)
	timeUint64, _ := ParseInt(string(resultMap["time"].(json.Number)), 10, 64)
	block.Time = time.Unix(timeUint64, 0)
	verUint64, _ := ParseUint(string(resultMap["version"].(json.Number)), 10, 32)
	block.Ver = uint32(verUint64)
	nonceUint64, _ := ParseUint(string(resultMap["nonce"].(json.Number)), 10, 32)
	block.Nonce = uint32(nonceUint64)
	bitsUint64, _ := ParseUint(resultMap["bits"].(string), 16, 32)
	block.Bits = uint32(bitsUint64)

	//Parse rpc result to transactions
	txsResult, _ := resultMap["tx"].([]interface{})
	txs := make([]*ModelTx, 0)
	for _, txResult := range txsResult {
		tx := new(ModelTx)
		bytesBuff, _ := hex.DecodeString(txResult.(string))
		tx.Hash = ReverseBytes(bytesBuff)
		txs = append(txs, tx)
	}
	return txs, nil
}

func (s *ModelSpendItem) NewModelSpendItem(result string) ([]*btcwire.MsgTx, error) {
	fmt.Println(result)
	return nil, nil
}

func GetAddressBalance(trans *gorp.Transaction, address string) (*ModelAddressBalance, error) {
	balanceBuff := make([]*ModelAddressBalance, 1)
	_, err := trans.Select(&balanceBuff, "select * from balance where Address=?", address)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if len(balanceBuff) == 1 {
		return balanceBuff[0], nil
	} else if len(balanceBuff) > 1 {
		log.Error("Mutilple addresses found.")
		return nil, errors.New("Multiple addresses found")

	} else {
		b := new(ModelAddressBalance)
		b.Address = address
		b.Balance = 0
		trans.Insert(b)
		log.Debug("New balance record. Address:%s.", b.Address)
		return b, nil
	}

}

func NewSpendItemIntoDB(trans *gorp.Transaction, msgTx *btcwire.MsgTx, mtx *ModelTx) error {
	//Handle transaction output recursivly, insert new record to spenditem database.
	for idx, out := range msgTx.TxOut {
		s := new(ModelSpendItem)
		s.OutTxId = mtx.Id
		s.OutScript = out.PkScript
		s.Value = out.Value
		s.Index = int64(idx)
		trans.Insert(s)
		log.Debug("New spenditem into database. Output tx id:%d, value:%d, index:%d", s.OutTxId, s.Value, s.Index)
	}

	//Handle transaction input recursibly, update the previous transaction output record in spenditem database.
	var sBuff []*ModelSpendItem
	for _, in := range msgTx.TxIn {
		//Find the tx record id in transaction database, by transaction's hashid and output index.
		prevTxId, _ := trans.SelectInt("select Id from tx where Hash=?", in.PreviousOutPoint.Hash.Bytes())
		sBuff = make([]*ModelSpendItem, 1)
		_, err := trans.Select(&sBuff, "select * from spend_item where OutTxId=? and Index=?", prevTxId, in.PreviousOutPoint.Index)
		if err != nil {
			log.Error(err.Error())
		}

		//update the spenditem
		if len(sBuff) == 1 {
			sBuff[0].InTxId = mtx.Id
			sBuff[0].InScript = in.SignatureScript
			trans.Update(sBuff[0])
			log.Debug("Update spenditem Id:%d, OutTxId:%d. Set InTxId=%d.", sBuff[0].Id, sBuff[0].OutTxId, mtx.Id)
		} else if len(sBuff) > 1 {
			log.Error("Multiple outputs matched for input previous tx, OutTxId=%d, index=%d.", prevTxId, in.PreviousOutPoint.Index)
		} else {
			log.Error("No output found in database.")
		}
	}
	return nil
}
