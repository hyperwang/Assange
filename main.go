package main

import (
	. "Assange/btcassange"
	//"fmt"
	//"github.com/go-sql-driver/mysql"
)

func main() {
	log := GetLogger("Main", DEBUG)
	dbmap := InitDb(Config)
	InitTables(dbmap)
	bw, _ := NewBlkWalker("blk00000.dat")
	//	for {
	for i := 0; i < 500; i++ {
		blk, fname, offset, err := bw.Next()
		if err != nil {
			log.Error(err.Error())
			return
		}
		hdr, err := NewBlkHdrItem(blk, fname, offset)
		InsertBlkHdrItemDirect(dbmap, hdr)
	}
	for i := 0; i < 10; i++ {
		HandleOrphanBlkHdrItem(dbmap)
	}

}
