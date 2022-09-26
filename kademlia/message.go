package d7024e

import (
	"bytes"
	"encoding/gob"
	"log"
)

// TODO Refactor name from msg -> packet (application layer messages called packets...)
// NOTE encoding/gob requires struct fields to be exported.
type msg struct {
	Method  rpc_method
	Caller  Contact // Will need an address of where to respond to as each instance is only keeping an open port.
	Payload content
}

// Simply have candidates or the respective value. eeez.
//Basically just need method and candidates, or key,value, or contact. (single contact).

type content struct {
	PingPong   string     //Simply calling it for face value, ping pong.
	Key        [160]byte  // FindValue: Find value.
	Value      [160]byte  // Payload for findvalue
	Store      Tuple      // Store: Key, Value
	Candidates []Contact  // Store/FindNode/FindValue candidates if not found.
	FindNode   KademliaID // Caller Key(kademlia ID) -> return (IP, Node ID) tuple for eahc of the k nodes closets to the target id.
}

type Tuple struct {
	Key   KademliaID
	Value KademliaID
}

// Careful with this 'enumeration' method, simple integers will work aswell.
type rpc_method int

const (
	Ping rpc_method = iota
	Store
	FindNode
	FindValue
)

// Uncesseary function.
//func CreateMsg(self Contact, method rpc_method, payload content) msg {
//	return msg{self, method, payload}
//}

func encodeMsg(m msg) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		log.Panic("COULD NOT ENCODE MSG:", m, err)
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
		log.Println("COULD NOT DECODE INC MSG:", inp, err)
		return inc, err
	}
	return inc, nil
}
