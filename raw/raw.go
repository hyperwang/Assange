package raw

import (
	. "Assange/blockdata"
	. "Assange/logging"
	"fmt"
	"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
)

var _ = fmt.Println
var log = GetLogger("Raw", DEBUG)

func NewTxFromRaw(raw []byte) *ModelTx {
	tx, err := btcutil.NewTxFromBytes(raw)
	if err != nil {
		log.Error(err.Error())
	}
	msgTx := tx.MsgTx()
	modelTx := new(ModelTx)
	modelTx.Hash = tx.Sha().String()
	if msgTx.TxIn[0].PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
		modelTx.IsCoinbase = true
	} else {
		modelTx.IsCoinbase = false
	}
	return modelTx
}
