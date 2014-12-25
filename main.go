package main

import (
	. "Assange/btcassange"
	"fmt"
	//"github.com/go-sql-driver/mysql"
	"flag"
	"github.com/coopernurse/gorp"
	"io"
	"path"
)

var reblockFlag bool

func init() {
	const (
		defaultReblock = false
		usage          = "Reindex the block headers into DB."
	)
	flag.BoolVar(&reblockFlag, "reblock", defaultReblock, usage)
}

var trimorphanFlag bool

func init() {
	const (
		defaultTrimorphan = false
		usage             = "Trim orphans in block header DB."
	)
	flag.BoolVar(&trimorphanFlag, "trimorphan", defaultTrimorphan, usage)
}

func main() {
	//log := GetLogger("Main", DEBUG)
	dbmap := InitDb(Config)
	InitTables(dbmap)
	flag.Parse()
	if reblockFlag {
		Reblock(dbmap)
	}
	if trimorphanFlag {
		TrimOrphan(dbmap)
	}
}

func Reblock(dbmap *gorp.DbMap) {
	log := GetLogger("Main", DEBUG)
	flist, _ := GetBlkFileList(Config.Block_data_dir)
	fmt.Println(flist)
	for _, f := range flist {
		bw, _ := NewBlkWalker(path.Join(Config.Block_data_dir, f))
		for {
			//for i := 0; i < 3000; i++ {
			blk, fname, offset, err := bw.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Error(err.Error())
				return
			}
			hdr, err := NewBlkHdrItem(blk, fname, offset)
			if err != nil {
				log.Error(err.Error())
			}
			err = CheckInsertBlkHdrItem(dbmap, hdr)
			if err == ERRDB_DUP_BLK {
			}
			if err == ERRDB_PRE_NOT_FOUND {
				InsertBlkHdrBuffer(hdr)
			}
			HandleOrphanBlkHdrItem(dbmap)
		}
	}
}

func TrimOrphan(dbmap *gorp.DbMap) {
	for {
		//for i := 0; i < 1; i++ {
		HandleOrphanBlkHdrItem(dbmap)
	}
}
