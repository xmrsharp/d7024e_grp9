package kademlia

import (
	"bytes"
	"encoding/gob"
	"log"
)

type msg struct {
	Method  RPCMethod
	Caller  Contact // Will need an address of where to respond to as each instance is only keeping an open port.
	Payload content
}

type content struct {
	PingPong   string     // Simply calling it for face value, ping pong.
	Key        KademliaID // FindValue: key.
	Value      string     // FindValue: value.
	Store      Tuple      // Store: Key, Value
	Candidates []Contact  // Store/FindNode/FindValue candidates if not found.
	FindNode   KademliaID // FindNode: Target
}

type Tuple struct {
	Key   KademliaID
	Value string
}

type RPCMethod int

func (r RPCMethod) String() string {
	switch r {
	case 0:
		return "PING"
	case 1:
		return "STORE"
	case 2:
		return "FIND_NODE"
	case 3:
		return "FIND_VALUE"
	default:
		return "UNKNOWN"
	}

}

const (
	Ping RPCMethod = iota
	Store
	FindNode
	FindValue
)

func encodeMsg(m msg) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		log.Printf(("COULD NOT ENCODE MSG: [%s] - %s"), m, err)
		return nil, nil
	}
	msg_buf := make([]byte, 2048)
	n, _ := buf.Read(msg_buf)
	return msg_buf[:n], nil
}

func decodeMsg(inp []byte) (m msg, e error) {
	buf := bytes.NewBuffer(inp)
	dec := gob.NewDecoder(buf)
	var inc msg
	err := dec.Decode(&inc)
	if err != nil {
		log.Printf(("COULD NOT DECODE MSG: [%s] - %s"), inp, err)
		return inc, err
	}
	return inc, nil
}
