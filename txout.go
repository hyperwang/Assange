package Assange

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	//"github.com/conformal/btcwire"
)

const (
	COINBASE = iota
	COMMONOUT
)

type TxOutItem struct {
	Id       int64
	Type     int
	Addr     []byte
	Value    int64
	Index    int
	TxHash   []byte
	BlkHash  []byte
	Spent    bool
	Expired  bool
	Orphaned bool
	Double   bool
}

type TxInItem struct {
	Id          int64
	TxHash      []byte
	BlkHash     []byte
	PreTxHash   []byte
	PreOutIndex int
	Expired     bool
	Orphaned    bool
}

func SplitTx(data []byte) ([]TxOutItem, []TxInItem, error) {
	var txOutSet []TxOutItem
	var txInSet []TxInItem
	var cbFlag bool
	txn, _ := btcutil.NewTxFromBytes(data)
	txnHash := txn.Sha().Bytes()
	mtxn := txn.MsgTx()

	if mtxn.TxIn[0].PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
		cbFlag = true
	} else {
		cbFlag = false
	}
	if !cbFlag {
		for i := range mtxn.TxIn {
			//_, _, pubkey, _ := DecodeSigScript(mtxn.TxIn[i].SignatureScript)
			//addr := btcutil.Hash160(pubkey)
			fmt.Println(hex.EncodeToString(mtxn.TxIn[i].SignatureScript))
			ti := TxInItem{
				TxHash:      txnHash,
				PreTxHash:   mtxn.TxIn[i].PreviousOutPoint.Hash.Bytes(),
				PreOutIndex: int(mtxn.TxIn[i].PreviousOutPoint.Index),
			}
			txInSet = append(txInSet, ti)
		}
	}

	if cbFlag && len(mtxn.TxOut) != 1 {
		return nil, nil, errors.New("Coinbase transaction should have only one transaction")
	}

	for i := range mtxn.TxOut {
		var txType int
		if cbFlag {
			txType = COINBASE
		} else {
			txType = COMMONOUT
		}

		//parse address
		_, addresses, _, err := btcscript.ExtractPkScriptAddrs(
			mtxn.TxOut[i].PkScript, &btcnet.MainNetParams)
		if err != nil {
			fmt.Println(err)
		}

		//make new TxOutItem
		to := TxOutItem{
			Type:   txType,
			Addr:   addresses[0].ScriptAddress(),
			Value:  mtxn.TxOut[i].Value,
			Index:  i,
			TxHash: txnHash,
		}
		txOutSet = append(txOutSet, to)
	}
	return txOutSet, txInSet, nil
}
