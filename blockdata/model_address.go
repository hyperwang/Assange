package blockdata

import (
	//. "Assange/util"
	//"encoding/hex"
	//"encoding/json"
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

type ModelAddress struct {
	Id      int64
	Address string
	Balance int64
}

type RelationTxoutAddress struct {
	Id        int64
	TxoutId   int64
	AddressId int64
}

func (a *ModelAddress) NewFromString(s string) {
	a.Address = s
	a.Balance = 0
}

func (a *ModelAddress) UpdateFromDbByAddress(trans *gorp.Transaction, address string) {
	addrBuff := new(ModelAddress)
	trans.SelectOne(&addrBuff, "select * from address where Address=?", address)
	if addrBuff.Id == 0 {
		a.Address = address
		a.Balance = 0
		err := trans.Insert(a)
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		a.Id = addrBuff.Id
		a.Address = addrBuff.Address
		a.Balance = addrBuff.Balance
	}
}

func (r *RelationTxoutAddress) InsertIntoDb(trans *gorp.Transaction, txout *ModelTxout, address *ModelAddress) {
	r.TxoutId = txout.Id
	r.AddressId = address.Id
	trans.Insert(r)
}

func InitModelAddress(dbmap *gorp.DbMap) {
	dbmap.AddTableWithName(ModelAddress{}, "address").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index uidx_address_address on address(Address)")

	dbmap.AddTableWithName(RelationTxoutAddress{}, "txoutaddress").SetKeys(true, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		log.Error(err.Error())
	}
	dbmap.Exec("create unique index idx_txoutaddress_txoutid_addressid on txoutaddress(TxoutId,AddressId)")
	dbmap.Exec("create index idx_txoutaddress_txoutid on txoutaddress(TxoutId)")
	dbmap.Exec("create index idx_txoutaddress_addressid on txoutaddress(AddressId)")

}
