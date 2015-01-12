package blockdata

import (
	//"Assange/config"
	//. "Assange/logging"
	//. "Assange/util"
	//"database/sql"
	//"encoding/hex"
	//"encoding/json"
	//"github.com/conformal/btcutil"
	//"errors"
	"github.com/conformal/btcscript"
	//"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	//. "strconv"
	//"time"
)

type ModelTxout struct {
	Id int64

	//Transaction output info
	OutTxHash string
	OutScript []byte
	OutIndex  int64
	Value     int64

	//Info extracted from script
	Type   btcscript.ScriptClass
	ReqSig int

	//More flags to be added
	IsCoinbase bool
	Extracted  bool
	Spent      bool

	//Transaction input
	RefTxinId int64
}

type ModelTxoutSet struct {
	TxOutSet []*ModelTxout
}

func (out *ModelTxout) NewFromUnextracted(trans *gorp.Transaction) error {
	err := trans.SelectOne(out, "select * from txout where Extracted=0 limit 1")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Uncalculated txout found. Id:%d.", out.Id)
	return nil
}

func (out *ModelTxout) NewFromMsg(msg_tx *btcwire.MsgTx, msg_txout *btcwire.TxOut, idx_txout int64) error {
	hash, err := msg_tx.TxSha()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	out.OutTxHash = hash.String()
	out.OutScript = msg_txout.PkScript
	out.Value = msg_txout.Value
	out.OutIndex = idx_txout
	out.Extracted = false
	out.Spent = false
	return nil
}

func (out *ModelTxout) InsertIntoDb(trans *gorp.Transaction) error {
	err := trans.Insert(out)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug("Insert new txout into DB. Output tx hash:%s, value:%d, index:%d.", out.OutTxHash, out.Value, out.OutIndex)
	return nil
}

func (outs *ModelTxoutSet) NewFromTx(tx *ModelTx) {
	for idx, msg_txout := range tx.Msg.TxOut {
		out := new(ModelTxout)
		out.NewFromMsg(tx.Msg, msg_txout, int64(idx))
		out.IsCoinbase = tx.IsCoinbase
		outs.TxOutSet = append(outs.TxOutSet, out)
	}
}

func (outs *ModelTxoutSet) InsertIntoDb(trans *gorp.Transaction) error {
	for _, out := range outs.TxOutSet {
		err := out.InsertIntoDb(trans)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitModelTxoutTable(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(ModelTxout{}, "txout").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index uidx_txout_outtxhash_outindex on txout(OutTxHash,OutIndex)")
	dbmap.Exec("create index idx_txout_extraced on txout(extracted)")
	dbmap.Exec("create index idx_txout_spent on txout(spent)")
}
