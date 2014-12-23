package btcassange

import (
	"encoding/binary"
	//"encoding/hex"
	"errors"
	"fmt"
	"github.com/conformal/btcutil"
	"os"
	//"time"
)

type BlkHdrItem struct {
	Id      int64
	Height  int64
	Hash    []byte
	PreHash []byte
	//Time     time.Time
	Bits     uint32
	Nonce    uint32
	Orphaned bool
}

func NewBlkHdrItem(blk *btcutil.Block) (*BlkHdrItem, error) {
	var hdrItem BlkHdrItem
	msgBlk := blk.MsgBlock()
	sha, _ := blk.Sha()

	//Block height
	hdrItem.Height = -1

	//Block hash
	hdrItem.Hash = sha.Bytes()
	//fmt.Println(hex.EncodeToString(hdrItem.Hash))

	//Previous block hash
	hdrItem.PreHash = msgBlk.Header.PrevBlock.Bytes()
	//fmt.Println(hex.EncodeToString(hdrItem.PreHash))

	//Time
	//hdrItem.Time = msgBlk.Header.Timestamp
	//fmt.Println(hdrItem.Time)

	//Bits
	hdrItem.Bits = msgBlk.Header.Bits
	//fmt.Println(hdrItem.Bits)

	//Nonce
	hdrItem.Nonce = msgBlk.Header.Nonce
	//fmt.Println(hdrItem.Nonce)

	//Orphaned flg
	hdrItem.Orphaned = true

	return &hdrItem, nil
}

type BlkWalker struct {
	fname      string
	f          *os.File
	nextOffset int64
}

func NewBlkWalker(s string) (*BlkWalker, error) {
	f, err := os.Open(s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	b := new(BlkWalker)
	b.fname = s
	b.f = f
	b.nextOffset = 0
	return b, nil
}

func (b *BlkWalker) Next() (*btcutil.Block, error) {
	var log = GetLogger("Block", DEBUG)
	var offset int64 = b.nextOffset
	int32buf := make([]byte, 4)

	n, err := b.f.Read(int32buf)
	if err != nil || n != len(int32buf) {
		return nil, err
	}
	offset += 4

	magic := binary.BigEndian.Uint32(int32buf)
	if magic != 0xf9beb4d9 {
		return nil, errors.New("Magic number error.")
	}

	_, err = b.f.Read(int32buf)
	offset += 4
	if err != nil || n != len(int32buf) {
		return nil, err
	}
	blkSize := binary.LittleEndian.Uint32(int32buf)

	blkBuff := make([]byte, blkSize)
	n, err = b.f.Read(blkBuff)
	if err != nil || n != len(blkBuff) {
		return nil, errors.New("Not read all the block data.")
	}
	offset += int64(blkSize)
	blk, _ := btcutil.NewBlockFromBytes(blkBuff)
	sha, _ := blk.Sha()
	log.Debug("Load block:%s from %s", sha.String(), b.fname)

	////Read 80 bytes for block header
	//hdrBuff := make([]byte, 80)
	//n, err = b.f.Read(hdrBuff)
	//if err != nil || n != len(hdrBuff) {
	//	return nil, err
	//}

	//Seek to the position of the next block
	n1, err := b.f.Seek(offset, 0)
	if err != nil || n1 != offset {
		return nil, errors.New("No seek to the right position.")
	}

	b.nextOffset = offset
	//blkHdr := BlkHdrItem{
	//	PreHash: ReverseBytes(hdrBuff[4:36]),
	//	Hash:    ReverseBytes(DSha256(hdrBuff[:])),
	//	Time:    binary.LittleEndian.Uint32(hdrBuff[68:72]),
	//	Bits:    binary.LittleEndian.Uint32(hdrBuff[72:76]),
	//	Nonce:   binary.LittleEndian.Uint32(hdrBuff[76:80]),
	//}
	//log.Debug("Previous block hash:%s", hex.EncodeToString(blkHdr.PreHash))
	//log.Debug("Current block hash:%s", hex.EncodeToString(blkHdr.Hash))
	return blk, nil
}

func ReverseBytes(b []byte) []byte {
	r := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = b[len(b)-1-i]
	}
	return r
}
