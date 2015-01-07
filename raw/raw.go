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
	modelTx := new(ModelTx)
	modelTx.Hash = tx.Sha().String()
	modelTx.IsCoinbase = false
	msgTx := tx.MsgTx()
	modelTx.Txouts = NewTxoutsFromMsg(msgTx, modelTx)
	modelTx.Txins = NewTxinsFromMsg(msgTx, modelTx)
	return modelTx
}

func NewBlockFromRaw(raw []byte) *ModelBlock {
	block, err := btcutil.NewBlockFromBytes(raw)
	if err != nil {
		log.Error(err.Error())
	}
	msgBlock := block.MsgBlock()
	modelBlock := new(ModelBlock)
	modelBlock.Height = block.Height()
	if hash, err := block.Sha(); err != nil {
		log.Error(err.Error())
	} else {
		modelBlock.Hash = hash.String()
	}
	modelBlock.PrevHash = msgBlock.Header.PrevBlock.String()
	modelBlock.MerkleRoot = msgBlock.Header.MerkleRoot.String()
	modelBlock.Time = msgBlock.Header.Timestamp
	modelBlock.Ver = msgBlock.Header.Version
	modelBlock.Nonce = msgBlock.Header.Nonce
	modelBlock.Bits = msgBlock.Header.Bits

	var tx *ModelTx
	for idx, msgTx := range msgBlock.Transactions {
		tx = new(ModelTx)
		hash, err := msgTx.TxSha()
		if err != nil {
			log.Error(err.Error())
		} else {
			tx.Hash = hash.String()
		}
		if idx == 0 {
			tx.IsCoinbase = true
		} else {
			tx.IsCoinbase = false
		}
		modelBlock.Txs = append(modelBlock.Txs, tx)
		modelBlock.Transactions = append(modelBlock.Transactions, hash.Bytes()...)
	}
	return modelBlock
}
