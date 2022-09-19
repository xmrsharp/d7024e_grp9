package network

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Have to keep attributes exported as encoding/gob package requires exported struct fields
type msg struct {
	Method  rpc_method
	Payload string
}

// Careful with this 'enumeration' method, simple integers will work aswell.
type rpc_method int

const (
	Ping rpc_method = iota
	Store
	FindData
	FindNode
)

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
