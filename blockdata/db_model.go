package blockdata

import (
	"Assange/config"
	. "Assange/logging"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
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
	Id      int64
	BlockId int64

	//Transaction info
	Hash []byte

	//More flags to be added
}

type ModelTxout struct {
	Id   int64
	TxId int64

	//Transaction output info
	Type  int64
	Addr  []byte
	Value int64
	Index int64

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
	dbmap.AddTableWithName(ModelTxout{}, "txout").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelTxin{}, "txin").SetKeys(true, "Id")
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

func NewBlockIntoDB(dbmap *gorp.DbMap, block *ModelBlock, tx []*ModelTx) error {
	trans, _ := dbmap.Begin()
	trans.Insert(block)
	log.Info("Block Id:%d", block.Id)
	for idx, _ := range tx {
		tx[idx].BlockId = block.Id
		trans.Insert(tx[idx])
	}
	trans.Commit()
	return nil
}
