package btcassange

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
)

var (
	ERRDB_BLK_NOT_FOUND = errors.New("Block not found in DB.")
	ERRDB_PRE_NOT_FOUND = errors.New("Previous block not found in DB.")
	ERRDB_DUP_BLK       = errors.New("Duplicated block not found in DB.")
)

func InitDb(config Configuration) *gorp.DbMap {
	var log = GetLogger("DB", DEBUG)
	source := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=True", config.Db_user, config.Db_password, config.Db_host, config.Db_database)
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
	//var log = GetLogger("DB", DEBUG)
	return nil
}

//Check database before inserting block header
//1. Check whether there is a block has the same hash, if not, return duplicatgin error.
//2. Check whether there is a block matchs previous hash, if not, return not on chain error.
func CheckInsertBlkHdrItem(dbmap *gorp.DbMap, blkHdr *BlkHdrItem) error {
	var log = GetLogger("DB", DEBUG)
	if IsBlkHdrInDb(dbmap, blkHdr) {
		return ERRDB_DUP_BLK
	}

	//Insert Genesis block directly.
	if hex.EncodeToString(blkHdr.Hash) == "6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000" {
		blkHdr.Height = 0
		blkHdr.Orphaned = false
		blkHdr.Tip = true
		log.Debug("Intend to insert Genesis block.")
		dbmap.Insert(blkHdr)
		log.Info("Insert Genesis block:%s into DB.", hex.EncodeToString(ReverseBytes(blkHdr.Hash)))
		return nil
	}

	preBlkHdr, err := GetPreBlkHdrFromDb(dbmap, blkHdr)
	if err != nil {
		return err
	}

	//Batch operations. Update tip and height
	trans, err := dbmap.Begin()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	preBlkHdr.Tip = false
	blkHdr.Height = preBlkHdr.Height + 1
	blkHdr.Tip = true
	blkHdr.Orphaned = preBlkHdr.Orphaned
	trans.Insert(blkHdr)
	trans.Update(preBlkHdr)
	err = trans.Commit()
	if err == nil {
		log.Info("Insert block(%s) into DB.",
			hex.EncodeToString(ReverseBytes(blkHdr.Hash)))
	}
	return err
}

func IsBlkHdrInDb(dbmap *gorp.DbMap, blkHdr *BlkHdrItem) bool {
	var log = GetLogger("DB", DEBUG)
	var blkHdrs []BlkHdrItem
	_, err := dbmap.Select(&blkHdrs, "select * from BlkHdrItem where Hash=?", blkHdr.Hash)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if len(blkHdrs) == 0 {
		return false
	}
	log.Warning("Duplicated block header found. Drop block:%s",
		hex.EncodeToString(blkHdr.Hash))
	return true
}

func GetPreBlkHdrFromDb(dbmap *gorp.DbMap, blkHdr *BlkHdrItem) (*BlkHdrItem, error) {
	var log = GetLogger("DB", DEBUG)
	var blkHdrs []BlkHdrItem
	_, err := dbmap.Select(&blkHdrs, "select * from BlkHdrItem where Hash=?", blkHdr.PreHash)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if len(blkHdrs) == 0 {
		log.Info("Block(%s)'s previous block(%s) not found in DB.",
			hex.EncodeToString(ReverseBytes(blkHdr.Hash)),
			hex.EncodeToString(ReverseBytes(blkHdr.PreHash)))
		return nil, ERRDB_PRE_NOT_FOUND
	}
	return &blkHdrs[0], nil
}

func HandleOrphanBlkHdrItem(dbmap *gorp.DbMap) {
	var log = GetLogger("DB", DEBUG)
	var idx int
	var hdr *BlkHdrItem
	if len(BlkHdrBuffer) == 0 {
		return
	}
	oldLen := len(BlkHdrBuffer)

	//Find a block in buffer randomly, and find its root recursively.
	seedIdx := rand.Intn(len(BlkHdrBuffer))
	for {
		idx, hdr = FindBlkHdrByHash(BlkHdrBuffer[seedIdx].PreHash)
		if hdr == nil {
			break
		}
		seedIdx = idx
	}

	idx = seedIdx
	//Try to insert it into DB.
	for {
		err := CheckInsertBlkHdrItem(dbmap, BlkHdrBuffer[idx])
		if err == ERRDB_DUP_BLK {
			log.Debug("Handling blocks done. Remain %d(%d).", oldLen, len(BlkHdrBuffer))
			return
		} else if err == nil {
			//Remove from buffer
			len_old := len(BlkHdrBuffer)
			currentHash := BlkHdrBuffer[idx].Hash
			BlkHdrBuffer = append(BlkHdrBuffer[:idx], BlkHdrBuffer[idx+1:]...)
			log.Info("Remove item from buffer. Previous length:%d, Current length:%d", len_old, len(BlkHdrBuffer))
			idx, hdr = FindBlkHdrByPreHash(currentHash)
			if hdr == nil {
				log.Debug("Handling blocks done. Remain %d(%d).", oldLen, len(BlkHdrBuffer))
				return
			}
		} else {
			log.Debug("Handling blocks done. Remain %d(%d).", oldLen, len(BlkHdrBuffer))
			return
		}
	}

	////Find a block in buffer randomly, and find its root recursively.
	////Try to insert it into DB.
	//for idx, _ := range BlkHdrBuffer {
	//	err := CheckInsertBlkHdrItem(dbmap, BlkHdrBuffer[idx])
	//	if err == ERRDB_DUP_BLK {
	//		return
	//	} else if err == nil {
	//		//Remove from buffer
	//		len_old := len(BlkHdrBuffer)
	//		BlkHdrBuffer = append(BlkHdrBuffer[:idx], BlkHdrBuffer[idx+1:]...)
	//		log.Info("Remove item from buffer. Previous length:%d, Current length:%d", len_old, len(BlkHdrBuffer))
	//	} else {
	//		return
	//	}
	//}
}

//func HandleOrphanBlkHdrItem(dbmap *gorp.DbMap) error {
//	var log = GetLogger("DB", DEBUG)
//	var blkHdrs1 []BlkHdrItem
//	//var blkHdrs2 []*BlkHdrItem
//	//var blkHdr BlkHdrItem
//	_, err := dbmap.Select(&blkHdrs1, "select * from BlkHdrItem where Orphaned=1 order by Time limit 200")
//	if err != nil {
//		log.Error(err.Error())
//		return err
//	}
//	for i := range blkHdrs1 {
//		log.Debug("Handle orphan block. Block hash:%s, Prevblock hash:%s",
//			hex.EncodeToString(ReverseBytes(blkHdrs1[i].Hash)),
//			hex.EncodeToString(ReverseBytes(blkHdrs1[i].PreHash)))
//		blkHdrs2, err := dbmap.Select(BlkHdrItem{}, "select * from BlkHdrItem where Hash=?", blkHdrs1[i].PreHash)
//		if err != nil {
//			log.Error(err.Error())
//			return err
//		}
//		if len(blkHdrs2) == 0 {
//			log.Warning("No previous block found. Set the block orphaned.")
//		} else if len(blkHdrs2) > 1 {
//			for j := range blkHdrs2 {
//				log.Debug("Previous blocks id:%d hash:%s",
//					blkHdrs2[j].(*BlkHdrItem).Id,
//					hex.EncodeToString(ReverseBytes(blkHdrs2[j].(*BlkHdrItem).Hash)))
//			}
//			log.Warning("More than one previous blocks found.")
//		} else {
//			if blkHdrs2[0].(*BlkHdrItem).Orphaned {
//				log.Warning("Based on an orphaned block. Set the block orphaned.")
//			} else {
//				blkHdrs1[i].Height = blkHdrs2[0].(*BlkHdrItem).Height + 1
//				blkHdrs1[i].Orphaned = false
//				_, err := dbmap.Update(&blkHdrs1[i])
//				if err != nil {
//					log.Error(err.Error())
//					return err
//				}
//				log.Info("Previous block found. Set height to %d.", blkHdrs1[i].Height)
//			}
//		}
//	}
//	return nil
//}
