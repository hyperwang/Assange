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
	//for {
	//	_, err := bw.Next()
	//	if err != nil {
	//		log.Error(err.Error())
	//	}
	//}
	//for {
	for i := 0; i < 2000; i++ {
		blk, err := bw.Next()
		if err != nil {
			log.Error(err.Error())
			return
		}
		hdr, err := NewBlkHdrItem(blk)
		InsertBlkHdrItem(dbmap, hdr)
		//HandleOrphanBlkHdrItem(dbmap)
	}
	for i := 0; i < 10; i++ {
		HandleOrphanBlkHdrItem(dbmap)
	}
}
