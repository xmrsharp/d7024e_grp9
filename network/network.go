package network

import (
	"log"
	"net"
	"strconv"
	"sync"
)

type Network struct {
	wg             sync.WaitGroup
	placeholder_id string
}

// TODO Constructor for server.
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

func (network *Network) handleRequest(m []byte, addr *net.UDPAddr) {
	// TODO Insert update call to channel to check if caller (addr) is existing in routing table and if not add to routing table.
	defer network.wg.Done()
	decode_msg, err := decodeMsg(m)
	if err != nil {
		// Simply want to end routine nicely.
		log.Println("UNKNOWN REQUEST BY:", addr)
		return
	}
	switch decode_msg.Method {
	case Ping:
		log.Println("RECIEVED PING REQUEST FROM:", addr)
		network.SendPingMessage(addr)
	case Store:
		log.Println("RECIEVED STORE REQUEST FROM:", addr)
	case FindData:
		log.Println("RECIEVED FIND_DATA REQUEST FROM:", addr)
	case FindNode:
		log.Println("RECIEVED FIND_NODE REQUEST FROM:", addr)
	default:
		log.Println("REQUEST FAILED BY UNKNOWN METHOD:", decode_msg.Method)
	}
}

func (network *Network) sendRequest(m msg, caller *net.UDPAddr) {
	payload, _ := encodeMsg(m)
	conn, err := net.DialUDP("udp4", nil, caller)
	if err != nil {
		log.Println("FAILED TO ESTABLISH CONNECTION TO: ", caller)
	}
	_, err = conn.Write(payload)
	if err != nil {
		log.Println("FAILED TO WRITE TO: ", caller.AddrPort())
	} else {
		log.Println("SENT REQ TO: ", caller)
	}

}

// TODO Change to take contact information when that part is complete.
// func (network *Network) SendPingMessage(contact *Contact) {
func (network *Network) SendPingMessage(addr *net.UDPAddr) {
	network.sendRequest(msg{Ping, "PONG"}, addr)
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

// Return wrapper of contact lists or data.
func (network *Network) SendFindDataMessage(hash string) {
	//TODO
}

// Return nothing as we're simply passing data to others to handle
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
