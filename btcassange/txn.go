package btcassange

import "github.com/conformal/btcutil"

import "encoding/hex"
import "github.com/conformal/btcscript"
import "github.com/conformal/btcnet"
import "fmt"
import "errors"
import "github.com/conformal/fastsha256"
import "hash"

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// DSha256 calculates the hash sha256(sha256(b)).
func DSha256(buf []byte) []byte {
	return calcHash(calcHash(buf, fastsha256.New()), fastsha256.New())
}

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

func DecodeSigScript(SignatureScript []byte) (fP2SH bool, Signature []byte, Pubkey []byte, err error) {
	sigLen := int(SignatureScript[0])
	pkLen := int(SignatureScript[sigLen+1])

	if sigLen+pkLen != len(SignatureScript)-2 {
		return false, nil, nil, errors.New("length error")
	}
	return btcscript.IsPayToScriptHash(SignatureScript), SignatureScript[1 : sigLen+1], SignatureScript[sigLen+2:], nil
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
	if !cbFlag {
		for i := range mtxn.TxIn {
			_, _, pubkey, _ := DecodeSigScript(mtxn.TxIn[i].SignatureScript)
			addr := btcutil.Hash160(pubkey)
			fmt.Println(hex.EncodeToString(mtxn.TxIn[i].SignatureScript))
			fmt.Println(hex.EncodeToString(addr))
			ti := TxnItem{
				Type:    0,
				TxnHash: txnHash,
				Addr:    addr,
			}
			txnItemSet = append(txnItemSet, ti)
		}
	}

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
