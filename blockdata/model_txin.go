package blockdata

import (
	//"Assange/config"
	//. "Assange/logging"
	//. "Assange/util"
	//"database/sql"
	//"encoding/hex"
	//"encoding/json"
	//"fmt"
	//"github.com/conformal/btcutil"
	//"errors"
	//"github.com/conformal/btcscript"
	//"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	//. "strconv"
	//"time"
)

type ModelTxin struct {
	Id int64

	//Transaction input info
	InTxHash     string
	InScript     []byte
	Sequence     uint32
	PrevOutHash  string
	PrevOutIndex int64

	//More flags to be added
	IsCoinbase bool
	Calculated bool
}

type ModelTxinSet struct {
	TxInSet []*ModelTxin
}

func (in *ModelTxin) NewFromMsg(msg_tx *btcwire.MsgTx, msg_txin *btcwire.TxIn) error {
	hash, err := msg_tx.TxSha()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	in.InTxHash = hash.String()
	in.InScript = msg_txin.SignatureScript
	in.Sequence = msg_txin.Sequence
	in.PrevOutHash = msg_txin.PreviousOutPoint.Hash.String()
	in.PrevOutIndex = int64(msg_txin.PreviousOutPoint.Index)
	in.Calculated = false
	return nil
}

func (in *ModelTxin) NewFromUncalculated(trans *gorp.Transaction) error {
	err := trans.SelectOne(in, "select * from txin where Calculated=0 limit 1")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Uncalculated txin found. Id:%d, TxHash:%s.", in.Id, in.InTxHash)
	return nil
}

func (in *ModelTxin) InsertIntoDb(trans *gorp.Transaction) error {
	//Insert txin into database
	err := trans.Insert(in)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Insert txin into DB. Id:%d", in.Id)
	return nil
}

func (ins *ModelTxinSet) NewFromTx(tx *ModelTx) {
	for _, msg_txin := range tx.Msg.TxIn {
		in := new(ModelTxin)
		in.NewFromMsg(tx.Msg, msg_txin)
		in.IsCoinbase = tx.IsCoinbase
		ins.TxInSet = append(ins.TxInSet, in)
	}
}

func (ins *ModelTxinSet) InsertIntoDb(trans *gorp.Transaction) error {
	for _, in := range ins.TxInSet {
		err := in.InsertIntoDb(trans)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitModelTxinTable(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(ModelTxin{}, "txin").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create index idx_txin_preouthash_preoutindex on txin(PreOutHash,PreOutIndex)")
	dbmap.Exec("create index idx_txin_intxhash on txin(InTxHash)")
	dbmap.Exec("create index idx_txin_calculated on txin(Calculated)")
}
