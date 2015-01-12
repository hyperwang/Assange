package zmq

import (
	//. "Assange/blockdata"
	//	. "Assange/raw"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	//"github.com/go-sql-driver/mysql"
	"github.com/coopernurse/gorp"
)

var _ = fmt.Println
var topic1 = "BLK"
var topic2 = "TXN"
var topic_len = len(topic1)
var socket *zmq.Socket
var dbmap *gorp.DbMap

func InitZmq(db *gorp.DbMap) {
	context, _ := zmq.NewContext()
	socket, _ = context.NewSocket(zmq.SUB)
	socket.Connect("tcp://127.0.0.1:5000")

	socket.SetSockOptString(zmq.SUBSCRIBE, topic1)
	socket.SetSockOptString(zmq.SUBSCRIBE, topic2)
	dbmap = db
}

func HandleZmq() {
	for {
		msg, _ := socket.Recv(0)
		topic := string(msg[0:topic_len])
		content := msg[topic_len:]
		if topic == "BLK" {
			HandleBlk(content)
		} else if topic == "TXN" {
			//HandleTxn(content)
		}
	}
}

func HandleTxn(raw []byte) {
	//trans, _ := dbmap.Begin()
	//tx := NewTxFromRaw(raw)
	//tx = InsertTxIntoDB(trans, tx)
	//for _, txout := range tx.Txouts {
	//	InsertTxoutIntoDb(trans, txout)
	//}
	//for _, txin := range tx.Txins {
	//	InsertTxinIntoDb(trans, txin)
	//}
	//trans.Commit()
}

func HandleBlk(raw []byte) {
	//trans, _ := dbmap.Begin()
	//block := NewBlockFromRaw(raw)
	//InsertBlockOnlyIntoDb(trans, block)
	//trans.Commit()
}
