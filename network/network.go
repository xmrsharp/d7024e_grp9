package network

import (
	"bytes"
	"encoding/gob"
	"log"
    "net"
	"strconv"
	"sync"
)

type Network struct {
	wg             sync.WaitGroup
	placeholder_id string
	//TODO Add kademlia ID
}

// TODO constructor for server.
// TODO plan, if ip is not full -> assume that this is the first node (i.e. dont connect).
func (network *Network) Listen(ip string, port int) {
	addr := ip + ":" + strconv.Itoa(port)
	udp_addr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
    	log.Println(err)
	}
	server_socket, err := net.ListenUDP("udp4", udp_addr)
	if err != nil {
	    log.Println(err)
    }
	log.Println("SERVING ON:", udp_addr)
	defer network.wg.Wait()
	defer server_socket.Close()
	for {
		buff := make([]byte, 1024, 1024)
		n, caller_addr, err := server_socket.ReadFromUDP(buff)
		if err != nil {
			log.Println("FAILED TO READ SOCKET:", err)
		} else {
			network.wg.Add(1)
            go network.handleRequest(buff[:n], caller_addr)
		}
	}
}


func encodeMsg(m msg) ([]byte, error){
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(m)
    if err != nil{
        log.Panic("COULD NOT ENCODE MSG:", m, err)
        return nil, nil
    }
    msg_buf := make([]byte,2048)
    n, _ := buf.Read(msg_buf)
    return msg_buf[:n],nil
}

func decodeMsg(inp []byte) (m msg, e error){
    buf := bytes.NewBuffer(inp)
    dec := gob.NewDecoder(buf)
    var inc msg
    err := dec.Decode(&inc)
    if err != nil{
        log.Println("COULD NOT DECODE INC MSG:", inp, err)
        return inc, err
    }
    return inc, nil
 }

func (network *Network) handleRequest(m []byte, addr *net.UDPAddr) {
	// TODO Check if addr in contact
	// TODO Check wether or not a response is desired (basically as mentioned above, create some generall udp message struct.)
	defer network.wg.Done()
    decode_msg, err := decodeMsg(m)
    if err != nil{
        // Simply want to end goroutine nicely.
        log.Println("REQUEST FAILED BY:", addr)
        return
    }
    switch decode_msg.Method{
    case Ping:
        // TODO
    case Store:
        // TODO 
    case FindData:
        // TODO
    case FindNode:
        // TODO
    default:
        log.Println("REQUEST FAILED BY UNKNOWN METHOD:", decode_msg.Method)

    }
    // For testing, assuming all incoming messages are encoded via func above.
    log.Println("Recieved from :", addr);
    log.Println("RECIEVED MSG:", decode_msg);
    ping_msg:= msg{FindNode, "TESTING"}
	log.Println("TEST msg:",ping_msg);
	test_encode, _ := encodeMsg(ping_msg)
    log.Println("ECODED MSG: ", test_encode);
	test_decode, _ := decodeMsg(test_encode);
	log.Println("DECODED MSG: ", test_decode);
}

func (network *Network) sendRequest() {
    
}

// Inc contant -> (ID, IP ADDRESS, DISTANCE.)
//func (network *Network) SendPingMessage(contact *Contact) {
// Simply ping ip address and append our contact info.
//
//}

//func (network *Network) SendFindContactMessage(contact *Contact) {
// TODO
//}

// Return wrapper of contact lists or data.
func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

// Return nothing as we're simply passing data to others to handle
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
