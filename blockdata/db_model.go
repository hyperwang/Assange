package blockdata

import (
	"Assange/config"
	. "Assange/logging"
	//. "Assange/util"
	"database/sql"
	//"encoding/hex"
	//"encoding/json"
	"fmt"
	//"github.com/conformal/btcutil"
	//"errors"
	//"github.com/conformal/btcscript"
	//"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	//. "strconv"
	//"time"
)

const (
	BytesPerBlockHash    = 32
	BytesPerBlockHashHex = BytesPerBlockHash * 2
)

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

func InitTables(dbmap *gorp.DbMap) {
	InitModelBlockTable(dbmap)
	InitModelTxTable(dbmap)
	InitModelTxoutTable(dbmap)
	InitModelTxinTable(dbmap)
	InitModelAddress(dbmap)
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

func GetMaxBlockIdFromDB(dbmap *gorp.DbMap) (int64, error) {
	var maxId int64
	maxId, _ = dbmap.SelectInt("select max(Id) as Id from block")
	return maxId, nil
}

func GetMaxTxIdFromDB(dbmap *gorp.DbMap) (int64, error) {
	var maxTxId int64
	maxTxId, _ = dbmap.SelectInt("select max(Id) as Id from tx")
	log.Info("Max id in transaction database is %d.", maxTxId)
	return maxTxId, nil
}

//func GetAddressBalance(trans *gorp.Transaction, address string) (*ModelBalance, error) {
//	var balanceBuff []*ModelBalance
//	oldLen := len(balanceBuff)
//	_, err := trans.Select(&balanceBuff, "select * from balance where Address=?", address)
//	if err != nil {
//		return nil, err
//	}
//	balanceBuff = balanceBuff[oldLen:len(balanceBuff)]
//
//	if len(balanceBuff) == 1 {
//		log.Debug("Address:%s found in database.", address)
//		return balanceBuff[0], nil
//	} else if len(balanceBuff) > 1 {
//		return nil, errors.New("Multiple addresses found")
//	} else {
//		b := new(ModelBalance)
//		b.Address = address
//		b.Balance = 0
//		trans.Insert(b)
//		log.Debug("New balance record. Address:%s.", b.Address)
//		return b, nil
//	}
//}

//func UpdateBalance(trans *gorp.Transaction, address string, value int64, flag bool) error {
//	b, err := GetAddressBalance(trans, address)
//	if err != nil {
//		return err
//	}
//	oldBalance := b.Balance
//	if flag {
//		b.Balance += value
//	} else {
//		b.Balance -= value
//	}
//	trans.Update(b)
//	log.Debug("Update address:%s balance from %d to %d.", b.Address, oldBalance, b.Balance)
//	return nil
//}
