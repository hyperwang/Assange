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
		usage          = "Reindex the block header into DB."
	)
	flag.BoolVar(&reblockFlag, "reblock", defaultReblock, usage)
}

func main() {
	//log := GetLogger("Main", DEBUG)
	dbmap := InitDb(Config)
	InitTables(dbmap)
	flag.Parse()
	if reblockFlag {
		Reblock(dbmap)
	}
}

func Reblock(dbmap *gorp.DbMap) {
	log := GetLogger("Main", DEBUG)
	flist, _ := GetBlkFileList(Config.Block_data_dir)
	fmt.Println(flist)
	for _, f := range flist {
		bw, _ := NewBlkWalker(path.Join(Config.Block_data_dir, f))
		for {
			//for i := 0; i < 10; i++ {
			blk, fname, offset, err := bw.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Error(err.Error())
				return
			}
			hdr, err := NewBlkHdrItem(blk, fname, offset)
			InsertBlkHdrItemDirect(dbmap, hdr)
		}
		for i := 0; i < 0; i++ {
			HandleOrphanBlkHdrItem(dbmap)
		}
	}
}
