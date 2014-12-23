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
	for i := 0; i < 2; i++ {
		blk, err := bw.Next()
		if err != nil {
			log.Error(err.Error())
		}
		hdr, err := NewBlkHdrItem(blk)
		dbmap.Insert(hdr)
	}
}
