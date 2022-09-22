package network

import (
	"log"
	"net"
	"strconv"
	"sync"
)

// For messages -> should only send KademliaID (as the address we get from the incoming socket, as of now atleast.).
// chan<- means send only. ch <- value
// <- chan means read only. value := <-ch
type Network struct {
	Wg       sync.WaitGroup
	addrs    *net.UDPAddr
	write_ch chan<- string
	read_ch  <-chan string
}

func InitNetwork(ip string, port int, write chan<- string, read <-chan string) *Network {
	udp_addr, err := net.ResolveUDPAddr("udp4", (ip + ":" + strconv.Itoa(port)))
	if err != nil {
		log.Panic("CANNOT SERVE ON SPECIFIED ADDR")
	}
	network := Network{sync.WaitGroup{}, udp_addr, write, read}
	return &network
}

func (network *Network) Listen() {
	server_socket, err := net.ListenUDP("udp4", network.addrs)
	if err != nil {
		log.Println(err)
	}
	log.Println("SERVING ON:", network.addrs)
	defer server_socket.Close()
	for {
		buff := make([]byte, 1024, 1024)
		n, caller_addr, err := server_socket.ReadFromUDP(buff)
		if err != nil {
			log.Println("FAILED TO READ SOCKET:", err)
		} else {
			network.Wg.Add(1)
			go network.handleRequest(buff[:n], caller_addr)
		}
	}
}

func (network *Network) handleRequest(m []byte, addr *net.UDPAddr) {
	// TODO Insert update call to channel to check if caller (addr) is existing in routing table and if not add to routing table.
	defer network.Wg.Done()
	decode_msg, err := decodeMsg(m)
	if err != nil {
		// Simply want to end routine nicely.
		log.Println("UNKNOWN REQUEST BY:", addr)
		return
	}
	// NOTE we need to examine payload as we're listening for all incoming connections.
	// And we need to check if payload is 0 -> then we know that we've been asked the question.
	// As if Payload is NOT 0 (can simply make a 0 check and return it to channel caller.) -> then we know its an incoming response. which tells us to simply return caller to
	// Atleast that is one way of doing it.
	// So basically plan -> implement channels here if payload is not 0.
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
		log.Println("SENT REQ: ", m.Method, " TO: ", caller)
	}
	network.write_ch <- "MESSAGE FROM GO ROUTINE"

}

// FOR ALL THE BELOW SEND MESSAGES WE WILL
// TODO Change to take contact information when that part is complete.
// func (network *Network) SendPingMessage(contact *Contact) {
func (network *Network) SendPingMessage(addr *net.UDPAddr) {
	m := new(msg)
	m.Method = Ping
	//network.sendRequest(msg{Ping, "PING"}, addr)
	network.sendRequest(*m, addr)
}

// func (network *Network) SendFindContactMessage(contact *Contact) {
// 	// TODO
// }

// Return wrapper of contact lists or data.
// func (network *Network) SendFindDataMessage(hash string) {
// 	//TODO
// }

// Return nothing as we're simply passing data to others to handle
// func (network *Network) SendStoreMessage(data []byte) {
// 	// TODO
// }
