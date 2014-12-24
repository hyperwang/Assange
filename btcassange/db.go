package btcassange

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

func InitDb(config Configuration) *gorp.DbMap {
	var log = GetLogger("DB", DEBUG)
	source := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", config.Db_user, config.Db_password, config.Db_host, config.Db_database)
	db, err := sql.Open("mysql", source)
	if err != nil {
		fmt.Println(err)
	}
	log.Info("Connect to database server:%s", config.Db_host)
	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
}

func InitTables(dbmap *gorp.DbMap) {
	dbmap.AddTable(TxInItem{}).SetKeys(true, "Id")
	dbmap.AddTable(TxOutItem{}).SetKeys(true, "Id")
	dbmap.AddTable(BlkHdrItem{}).SetKeys(true, "Id")
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		fmt.Println(err)
	}
}

func InsertBlkHdrItem(dbmap *gorp.DbMap, blkHdr *BlkHdrItem) error {
	var log = GetLogger("DB", DEBUG)
	if hex.EncodeToString(blkHdr.Hash) == "6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000" {
		var blkHdrs []BlkHdrItem
		_, err := dbmap.Select(&blkHdrs, "select * from BlkHdrItem where Height=0 and Orphaned=0")
		if err != nil {
			log.Error(err.Error())
			return err
		}
		if len(blkHdrs) == 0 {
			blkHdr.Height = 0
			blkHdr.Orphaned = false
			log.Debug("Genesis block found.")
		} else {
			log.Error("Genesis block already in DB.")
			return errors.New("Genesis block exists.")
		}
	} else {
		var blkHdrs []BlkHdrItem
		_, err := dbmap.Select(&blkHdrs, "select * from BlkHdrItem where Hash=?", blkHdr.PreHash)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		if len(blkHdrs) == 0 {
			blkHdr.Orphaned = true
			log.Warning("No previous block found. Set the block orphaned.")
		} else if len(blkHdrs) > 1 {
			log.Error("More than one previous blocks found. Drop the block.")
			return errors.New("More than one previous")
		} else {
			var blkHdrs1 []BlkHdrItem
			_, err := dbmap.Select(&blkHdrs1, "select * from BlkHdrItem where Hash=?", blkHdr.Hash)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			if len(blkHdrs1) != 0 {
				log.Error("Block exists.")
				return errors.New("Block exists.")
			}
			if blkHdrs[0].Orphaned {
				blkHdr.Height = blkHdrs[0].Height + 1
				blkHdr.Orphaned = true
				log.Warning("Based on an orphaned block. Set the block orphaned.")
			} else {
				blkHdr.Height = blkHdrs[0].Height + 1
				blkHdr.Orphaned = false
				log.Info("Previous block found. Set height to %d.", blkHdr.Height)
			}
		}
	}
	dbmap.Insert(blkHdr)
	return nil
}

func HandleOrphanBlkHdrItem(dbmap *gorp.DbMap) error {
	var log = GetLogger("DB", DEBUG)
	var blkHdrs1 []BlkHdrItem
	//var blkHdrs2 []*BlkHdrItem
	//var blkHdr BlkHdrItem
	_, err := dbmap.Select(&blkHdrs1, "select * from BlkHdrItem where Orphaned=1 order by Height")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	for i := range blkHdrs1 {
		log.Debug("Handle orphan block. Block hash:%s, Prevblock hash:%s",
			hex.EncodeToString(ReverseBytes(blkHdrs1[i].Hash)),
			hex.EncodeToString(ReverseBytes(blkHdrs1[i].PreHash)))
		blkHdrs2, err := dbmap.Select(BlkHdrItem{}, "select * from BlkHdrItem where Hash=?", blkHdrs1[i].PreHash)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		if len(blkHdrs2) == 0 {
			log.Warning("No previous block found. Set the block orphaned.")
		} else if len(blkHdrs2) > 1 {
			for j := range blkHdrs2 {
				log.Debug("Previous blocks id:%d hash:%s",
					blkHdrs2[j].(*BlkHdrItem).Id,
					hex.EncodeToString(ReverseBytes(blkHdrs2[j].(*BlkHdrItem).Hash)))
			}
			log.Warning("More than one previous blocks found.")
		} else {
			//fmt.Print(blkHdrs2, len(blkHdrs2), blkHdrs2[0].(*BlkHdrItem).Id)
			if blkHdrs2[0].(*BlkHdrItem).Orphaned {
				log.Warning("Based on an orphaned block. Set the block orphaned.")
			} else {
				blkHdrs1[i].Height = blkHdrs2[0].(*BlkHdrItem).Height + 1
				blkHdrs1[i].Orphaned = false
				_, err := dbmap.Update(&blkHdrs1[i])
				if err != nil {
					log.Error(err.Error())
					return err
				}
				log.Info("Previous block found. Set height to %d.", blkHdrs1[i].Height)
			}
		}
	}
	return nil
}
