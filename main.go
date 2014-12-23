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
	for i := 0; i < 500; i++ {
		blk, err := bw.Next()
		if err != nil {
			log.Error(err.Error())
			return
		}
		hdr, err := NewBlkHdrItem(blk)
		InsertBlkHdrItem(dbmap, hdr)
	}
}
