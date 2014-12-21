package Assange

import "github.com/conformal/btcutil"

import "encoding/hex"
import "github.com/conformal/btcscript"
import "github.com/conformal/btcnet"
import "fmt"

//import btcwire "github.com/conformal/btcwire"

/*func main() {
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.SUB)
	socket.SetSockOptString(zmq.SUBSCRIBE, blk_topic)
	socket.SetSockOptString(zmq.SUBSCRIBE, txn_topic)
	socket.Connect("tcp://127.0.0.1:5000")
	println(btcwire.MaxBlockPayload)
	for {
		msg, _ := socket.Recv(0)
		topic := string(msg[:3])
		data := msg[3:]
		if topic == "TXN" {
			println("--TXN--")
			txn, _ := btcutil.NewTxFromBytes(data)
			mtxn := txn.MsgTx()
			mtxn_sha, _ := mtxn.TxSha()
			println(mtxn_sha.String())
		}
	}
}*/

type TxnItem struct {
	Id        int64
	Type      int8 //0:common txin; 1:common txout; 2:coinbase txin; 3:coinbase txout
	TxnHash   []byte
	Addr      []byte
	Value     int64
	BlkHash   []byte
	Confirmed int64
}

func (t TxnItem) String() string {
	return fmt.Sprintf("Id:%d\nType:%d\nTxnHash:%s\nAddr:%s\nValue:%d\n", t.Id, t.Type, hex.EncodeToString(t.TxnHash), hex.EncodeToString(t.Addr), t.Value)
}

// Split the transaction into TxnItem slice
// fill the Type,TxnHash,Addr,Value(tx output) FIELDS
// leave the Id,BlkHash,Value(tx input),Confirmed FILEDS
func SplitTxnFromString(data []byte) []TxnItem {
	var txnItemSet []TxnItem
	var cbFlag bool
	txn, _ := btcutil.NewTxFromBytes(data)
	txnHash := txn.Sha().Bytes()
	//println(string(txnHash))
	mtxn := txn.MsgTx()

	//println("TxIn", len(mtxn.TxIn))
	if mtxn.TxIn[0].PreviousOutPoint.Hash.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
		cbFlag = true
	} else {
		cbFlag = false
	}
	//println("Coinbase:", cbFlag)

	//if is coinbase txn, skip the txin handling.
	//	if !cbFlag {
	//		for i := range mtxn.TxIn {
	//			ti := TxnItem{
	//				Type:    0,
	//				TxnHash: txnHash,
	//			}
	//		}
	//	}

	//println("TxOut", len(mtxn.TxOut))
	for i := range mtxn.TxOut {
		var txType int8
		if cbFlag {
			txType = 3
		} else {
			txType = 1
		}
		//println(txType)

		//parse address
		_, addresses, _, err := btcscript.ExtractPkScriptAddrs(
			mtxn.TxOut[i].PkScript, &btcnet.MainNetParams)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println("Script Class:", scriptClass)
		//fmt.Println("Addresses:", addresses)
		//fmt.Println("Required Signatures:", reqSigs)

		//make new TxnItem
		to := TxnItem{
			Type:    txType,
			TxnHash: txnHash,
			Addr:    addresses[0].ScriptAddress(),
			Value:   mtxn.TxOut[i].Value,
		}
		//fmt.Println(to)
		txnItemSet = append(txnItemSet, to)
	}
	//fmt.Printf("%v\n", to)
	return txnItemSet
}
