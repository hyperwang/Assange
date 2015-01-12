package blockdata

import (
	//	"Assange/config"
	//	. "Assange/logging"
	//	. "Assange/util"
	//	"database/sql"
	"encoding/hex"
	//"encoding/json"
	//"fmt"
	"github.com/conformal/btcutil"
	//	"errors"
	//	"github.com/conformal/btcscript"
	//	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	//. "strconv"
	"time"
)

type ModelTx struct {
	Id int64

	//Transaction info
	Hash         string
	Ver          int32
	LockTime     uint32
	ReceivedTime time.Time

	//More flags to be added
	IsCoinbase bool
	Extracted  bool
	Confirmed  bool

	Msg *btcwire.MsgTx `db:"-"`
}

type RelationBlockTx struct {
	Id int64

	//Including relation between block and tx
	BlockId int64
	TxId    int64

	//More flags to be added
}

func (tx *ModelTx) InsertIntoDb(trans *gorp.Transaction) error {
	err := trans.Insert(tx)
	if err != nil {
		return err
	}
	log.Info("Insert new transaction, Id:%d, Hash:%s.", tx.Id, tx.Hash)
	return nil
}

func (r *RelationBlockTx) InsertIntoDb(trans *gorp.Transaction, block *ModelBlock, tx *ModelTx) {
	r.BlockId = block.Id
	r.TxId = tx.Id
	trans.Insert(r)
}

func (tx *ModelTx) NewFromUnextracted(trans *gorp.Transaction) error {
	err := trans.SelectOne(tx, "select * from tx where Extracted=0 limit 1")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Unextracted tx found. Id:%d Hash:%s.", tx.Id, tx.Hash)
	return nil
}

func (tx *ModelTx) NewFromString(result string) {
	bytesResult, _ := hex.DecodeString(result)
	tx1, err := btcutil.NewTxFromBytes(bytesResult)
	if err != nil {
		log.Error(err.Error())
	}
	tx.Msg = tx1.MsgTx()
	hash, _ := tx.Msg.TxSha()
	tx.Hash = hash.String()
	tx.Ver = tx.Msg.Version
	tx.LockTime = tx.Msg.LockTime
	tx.Extracted = false
}

func (tx *ModelTx) UpdateInOutFromString(result string) {
	bytesResult, _ := hex.DecodeString(result)
	tx1, err := btcutil.NewTxFromBytes(bytesResult)
	if err != nil {
		log.Error(err.Error())
	}
	tx.Msg = tx1.MsgTx()
}

func InitModelTxTable(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(ModelTx{}, "tx").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index uidx_tx_hash on tx(Hash)")
	dbmap.Exec("create index idx_tx_blockhash on tx(BlockHash)")
	dbmap.Exec("create index idx_tx_extracted on tx(Extracted)")
	dbmap.Exec("create index idx_tx_confirmed on tx(Confirmed)")

	dbmap.AddTableWithName(RelationBlockTx{}, "blocktx").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index uidx_blocktx_blockid_txid on blocktx(BlockId,TxId)")
	dbmap.Exec("create index idx_blocktx_blockid on blocktx(BlockId)")
	dbmap.Exec("create index idx_blocktx_txid on blocktx(TxId)")
}
