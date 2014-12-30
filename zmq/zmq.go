package zmq

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

var topic1 = "BLK"
var topic2 = "TXN"

func InitZmq() {
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.SUB)
	socket.Connect("tcp://127.0.0.1:5000")

	socket.SetSockOptString(zmq.SUBSCRIBE, topic1)
	socket.SetSockOptString(zmq.SUBSCRIBE, topic2)
	for {
		msg, _ := socket.Recv(0)
		fmt.Println(msg)
	}
}
