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
	"github.com/conformal/btcutil"
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
	Hash       string
	PrevHash   string
	NextHash   string
	MerkleRoot string
	Time       time.Time
	Ver        int32
	Nonce      uint32
	Bits       uint32

	//Transactions
	Txs          []*ModelTx `db:"-"`
	Transactions []byte

	//More flags to be added
	Extracted bool
}

type ModelTx struct {
	Id int64

	//Transaction info
	Hash         string
	Ver          int32
	LockTime     uint32
	ReceivedTime time.Time

	//Txouts
	Txouts []*ModelTxout `db:"-"`

	//Txins
	Txins []*ModelTxin `db:"-"`

	//More flags to be added
	IsCoinbase bool
	Extracted  bool
	Confirmed  bool
}

type RelationBlockTx struct {
	Id int64

	//Including relation between block and tx
	BlockId int64
	TxId    int64

	//More flags to be added
}

type ModelBalance struct {
	Id int64

	Address string
	Balance int64
}

type ModelTxout struct {
	Id int64

	//Transaction output info
	OutTxHash string
	OutScript []byte
	OutIndex  int64
	Address   string
	Value     int64

	//More flags to be added
	IsCoinbase bool

	//Transaction input
	RefTxinId int64
}

type ModelTxin struct {
	Id int64

	//Transaction input info
	InTxHash     string
	InScript     []byte
	Sequence     uint32
	PrevOutHash  string
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
	dbmap.AddTableWithName(ModelTxout{}, "txout").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelTxin{}, "txin").SetKeys(true, "Id")
	dbmap.AddTableWithName(ModelBalance{}, "balance").SetKeys(true, "Id")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return err
	}

	dbmap.Exec("create unique index uidx_block_hash on block(Hash)")
	dbmap.Exec("alter table `block` add index `idx_block_prevhash` (PrevHash)")
	dbmap.Exec("alter table `block` add index `idx_block_nexthash` (NextHash)")
	dbmap.Exec("alter table `block` add index `idx_block_height` (Height)")

	dbmap.Exec("create unique index uidx_tx_hash on tx(Hash)")
	dbmap.Exec("alter table `tx` add index `idx_tx_blockhash` (BlockHash)")

	dbmap.Exec("create unique index uidx_txout_outtxhash_outindex on txout(OutTxHash,OutIndex)")
	dbmap.Exec("alter table `txout` add index `idx_txout_address` (Address)")

	dbmap.Exec("create index idx_txin_preouthash_preoutindex on txin(PreOutHash,PreOutIndex)")
	dbmap.Exec("alter table `txin` add index `idx_txin_intxhash` (InTxHash)")

	dbmap.Exec("create unique index uidx_balance_address on balance(Address)")

	return nil
}

func GetMaxBlockHeightFromDB(dbmap *gorp.DbMap) (int64, error) {
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

func InsertTxIntoDB(trans *gorp.Transaction, tx *ModelTx) *ModelTx {
	var txBuff []*ModelTx
	oldLen := len(txBuff)
	trans.Select(&txBuff, "select * from tx where Hash=?", tx.Hash)
	if oldLen == len(txBuff) {
		err := trans.Insert(tx)
		if err != nil {
			log.Error(err.Error())
		}
		return tx
	} else {
		log.Error("Tx hash already existed. Id:%d, Hash:%s.", tx.Id, tx.Hash)
		return txBuff[oldLen]
	}
}

func InsertBlockIntoDb(trans *gorp.Transaction, block *ModelBlock) {
	block.Extracted = true
	err := trans.Insert(block)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("Insert new block, Id:%d, Height:%d.", block.Id, block.Height)
	for idx, _ := range block.Txs {
		block.Txs[idx] = InsertTxIntoDB(trans, block.Txs[idx])
		blockTx := new(RelationBlockTx)
		blockTx.BlockId = block.Id
		blockTx.TxId = block.Txs[idx].Id
		err := trans.Insert(blockTx)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func InsertRelationBlockTxIntoDB(trans *gorp.Transaction, block *ModelBlock, tx *ModelTx) {
	r := new(RelationBlockTx)
	r.BlockId = block.Id
	r.TxId = tx.Id
	trans.Insert(r)
}

func InsertBlockOnlyIntoDb(trans *gorp.Transaction, block *ModelBlock) {
	var blockBuff []*ModelBlock

	oldLen := len(blockBuff)
	_, err := trans.Select(&blockBuff, "select * from block where Hash=?", block.PrevHash)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if oldLen == len(blockBuff) {
		log.Error("No previous block found. Hash:%s.", block.PrevHash)
		return
	} else {
		blockBuff[0].NextHash = block.Hash
		trans.Update(blockBuff[0])
	}

	err = trans.Insert(block)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info("Insert new block, Id:%d, Height:%d.", block.Id, block.Height)
	return
}

func GetOneUnextractedBlock(trans *gorp.Transaction) (*ModelBlock, error) {
	var blockBuff []*ModelBlock
	oldLen := len(blockBuff)
	_, err := trans.Select(&blockBuff, "select * from block where Extracted=0 limit 1")
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if oldLen == len(blockBuff) {
		log.Error("No unextracted block found.")
		return nil, errors.New("Unextracted block not found")
	} else {
		log.Debug("Unextracted block found. Id:%d Hash:%s.", blockBuff[0].Id, blockBuff[0].Hash)
	}
	return blockBuff[0], nil
}

func NewTxFromString(result string, tx *ModelTx) {
	bytesResult, _ := hex.DecodeString(result)
	tx1, err := btcutil.NewTxFromBytes(bytesResult)
	if err != nil {
		log.Error(err.Error())
	}
	msgTx := tx1.MsgTx()
	hash, _ := msgTx.TxSha()
	tx.Hash = hash.String()
	tx.Ver = msgTx.Version
	tx.LockTime = msgTx.LockTime
	tx.Extracted = false
	tx.Txouts = NewTxoutsFromMsg(msgTx, tx)
	tx.Txins = NewTxinsFromMsg(msgTx, tx)
}

func NewBlockFromMap(resultMap map[string]interface{}) (*ModelBlock, error) {
	block := new(ModelBlock)
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
		bHash, err := hex.DecodeString(tx.Hash)
		if err != nil {
			log.Error(err.Error())
		}
		block.Transactions = append(block.Transactions, ReverseBytes(bHash)...)
	}
	return block, nil
}

func GetAddressBalance(trans *gorp.Transaction, address string) (*ModelBalance, error) {
	var balanceBuff []*ModelBalance
	oldLen := len(balanceBuff)
	_, err := trans.Select(&balanceBuff, "select * from balance where Address=?", address)
	if err != nil {
		return nil, err
	}
	balanceBuff = balanceBuff[oldLen:len(balanceBuff)]

	if len(balanceBuff) == 1 {
		log.Debug("Address:%s found in database.", address)
		return balanceBuff[0], nil
	} else if len(balanceBuff) > 1 {
		return nil, errors.New("Multiple addresses found")
	} else {
		b := new(ModelBalance)
		b.Address = address
		b.Balance = 0
		trans.Insert(b)
		log.Debug("New balance record. Address:%s.", b.Address)
		return b, nil
	}
}

func NewTxoutsFromMsg(msgTx *btcwire.MsgTx, mtx *ModelTx) []*ModelTxout {
	var outBuff []*ModelTxout
	//Handle transaction output recursivly.
	for idx, out := range msgTx.TxOut {
		s := new(ModelTxout)
		hash, err := msgTx.TxSha()
		if err != nil {
			log.Error(err.Error())
		}
		s.OutTxHash = hash.String()
		s.IsCoinbase = mtx.IsCoinbase
		s.OutScript = out.PkScript
		s.Value = out.Value
		s.OutIndex = int64(idx)
		//Extract address
		s.Address = ExtractAddrFromScript(s.OutScript)
		outBuff = append(outBuff, s)
	}
	return outBuff
}

func NewTxinsFromMsg(msgTx *btcwire.MsgTx, mtx *ModelTx) []*ModelTxin {
	var inBuff []*ModelTxin
	if mtx.IsCoinbase {
		return inBuff
	}
	//Handle transaction input recursibly.
	for _, in := range msgTx.TxIn {
		s := new(ModelTxin)
		hash, err := msgTx.TxSha()
		if err != nil {
			log.Error(err.Error())
		}
		s.InTxHash = hash.String()
		s.InScript = in.SignatureScript
		s.Sequence = in.Sequence
		s.PrevOutHash = in.PreviousOutPoint.Hash.String()
		s.PrevOutIndex = int64(in.PreviousOutPoint.Index)
		inBuff = append(inBuff, s)
	}
	return inBuff
}

func InsertTxinIntoDb(trans *gorp.Transaction, txin *ModelTxin) {
	//Insert txin into database
	err := trans.Insert(txin)
	if err != nil {
		log.Error(err.Error())
		return
	}

	//Find matched txout, and update address balance
	var outBuff []*ModelTxout
	oldLen := len(outBuff)
	query := fmt.Sprintf("select * from txout where OutTxHash='%s' and OutIndex=%d", txin.PrevOutHash, txin.PrevOutIndex)
	_, err = trans.Select(&outBuff, query)
	if err != nil {
		log.Error(err.Error())
	}
	outBuff = outBuff[oldLen:len(outBuff)]

	if len(outBuff) == 1 {
		outBuff[0].RefTxinId = txin.Id
		trans.Update(outBuff[0])
		log.Debug("Update txout Id:%d, OutTxHash:%s, OutIndex:%d. Set RefTxinId=%d.", outBuff[0].Id, outBuff[0].OutTxHash, outBuff[0].OutIndex, outBuff[0].RefTxinId)
		UpdateBalance(trans, outBuff[0].Address, outBuff[0].Value, false)
	} else if len(outBuff) > 1 {
		log.Error("Multiple outputs matched for input previous tx, OutTxHash:%s, OutIndex=%d.", txin.PrevOutHash, txin.PrevOutIndex)
	} else {
		log.Error("No output found in database. OutTxHash=%s, OutIndex=%d.", txin.PrevOutHash, txin.PrevOutIndex)
	}
}

func InsertTxoutIntoDb(trans *gorp.Transaction, txout *ModelTxout) {
	err := trans.Insert(txout)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Debug("New spenditem into database. Output tx hash:%s, value:%d, index:%d, address:%s.", txout.OutTxHash, txout.Value, txout.OutIndex, txout.Address)
	UpdateBalance(trans, txout.Address, txout.Value, true)
}

func UpdateBalance(trans *gorp.Transaction, address string, value int64, flag bool) error {
	b, err := GetAddressBalance(trans, address)
	if err != nil {
		return err
	}
	oldBalance := b.Balance
	if flag {
		b.Balance += value
	} else {
		b.Balance -= value
	}
	trans.Update(b)
	log.Debug("Update address:%s balance from %d to %d.", b.Address, oldBalance, b.Balance)
	return nil
}
