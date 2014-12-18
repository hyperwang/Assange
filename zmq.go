package main

import zmq "github.com/alecthomas/gozmq"
import btcutil "github.com/conformal/btcutil"
import "github.com/conformal/btcwire"
//import "strings"

const blk_topic = "BLK"
const txn_topic = "TXN"

func main() {
  context, _ := zmq.NewContext()
  socket, _ := context.NewSocket(zmq.SUB)
  socket.SetSockOptString(zmq.SUBSCRIBE,blk_topic)
  socket.SetSockOptString(zmq.SUBSCRIBE,txn_topic)
  socket.Connect("tcp://127.0.0.1:5000")
  println(btcwire.MaxBlockPayload)
  for {
    msg, _ := socket.Recv(0)
    topic := string(msg[:3])
    data  := msg[3:]
    if topic == "TXN" {
      println("--TXN--")
      txn, _ := btcutil.NewTxFromBytes(data)
      mtxn := txn.MsgTx()
      mtxn_sha, _ := mtxn.TxSha()
      println(mtxn_sha.String()) 
    }
  }
}

