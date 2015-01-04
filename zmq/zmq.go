package zmq

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

var topic1 = "BLK"
var topic2 = "TXN"
var topic_len = len(topic1)
var socket *zmq.Socket

func InitZmq() {
	context, _ := zmq.NewContext()
	socket, _ = context.NewSocket(zmq.SUB)
	socket.Connect("tcp://127.0.0.1:5000")

	socket.SetSockOptString(zmq.SUBSCRIBE, topic1)
	socket.SetSockOptString(zmq.SUBSCRIBE, topic2)
}

func HandleZmq() {
	for {
		msg, _ := socket.Recv(0)
		topic := string(msg[0:topic_len])
		content := msg[topic_len:]
		if topic == "BLK" {
			HandleBlk(&content)
		} else if topic == "TXN" {
			HandleTxn(&content)
		}
	}
}

func HandleTxn(raw *[]byte) {
	fmt.Println("txn")
}

func HandleBlk(raw *[]byte) {
	fmt.Println("block")
}
