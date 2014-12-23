package Assange

import (
	//"fmt"
	"testing"
)

func TestNext_1(t *testing.T) {
	blk, err := NewBlkWalker("blk00000.dat")
	if err != nil {
		t.Error(err)
	}

	//fmt.Println(blk)
	blk.Next()
	blk.Next()
	t.Error("finally.")
}
