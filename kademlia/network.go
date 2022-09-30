package d7024e

import (
	"log"
	"net"
	"sync"
)

type Network struct {
	msgHeader  Contact
	wg         sync.WaitGroup
	outRequest *SharedMap
	addrs      *net.UDPAddr
	write_ch   chan<- msg
	read_ch    <-chan msg
}

func InitNetwork(msgHeader Contact, addrs string, outRequest *SharedMap, write chan<- msg, read <-chan msg) *Network {
	udp_addr, err := net.ResolveUDPAddr("udp4", addrs)
	if err != nil {
		log.Panic("CANNOT SERVE ON SPECIFIED ADDR")
	}
	network := Network{msgHeader, sync.WaitGroup{}, outRequest, udp_addr, write, read}
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
		// Change size of reader.
		buff := make([]byte, 4096, 4096)
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
	defer network.wg.Done()
	decode_msg, err := decodeMsg(m)
	if err != nil {
		// Simply want to end routine nicely.
		log.Println("UNKNOWN MSG BY:", addr)
		return
	}
	log.Println("HANDLE REQUEST - RECIEVED: ", decode_msg.Method, " REQUEST FROM: ", addr)
	network.write_ch <- decode_msg
}

// Add caller to awaiting result channel?
// Here need to alter the dialup to simply call the addrs from contact.
func (network *Network) sendRequest(m msg, to Contact) {
	payload, _ := encodeMsg(m)
	log.Println("ADDRESS:", to.Address)
	udp_addr, err := net.ResolveUDPAddr("udp4", to.Address)
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO ESTABLISH CONNECTION TO: ", to)
	}
	conn, err := net.DialUDP("udp4", nil, udp_addr)
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO ESTABLISH CONNECTION TO: ", to)
	}
	_, err = conn.Write(payload)
	defer conn.Close()
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO WRITE TO: ", udp_addr)
	} else {
		log.Println("SENT REQ: ", m.Method, " TO: ", to.Address)
		conn.Close()
	}

}

// TODO Refactor: awkward having to append self to every packet payload.
func (network *Network) SendPingMessage(to *Contact, pingMsg string) {
	m := new(msg)
	m.Method = Ping
	m.Caller = network.msgHeader
	m.Payload.PingPong = pingMsg
	network.sendRequest(*m, *to)
}

func (network *Network) respondFindContactMessage(to *Contact, candidates []Contact) {
	m := new(msg)
	m.Method = FindNode
	m.Caller = network.msgHeader
	m.Payload.Candidates = candidates
	network.outRequest.mutex.Lock()
	network.outRequest.outgoingRegister[*to.ID] += 1
	network.outRequest.mutex.Unlock()
	network.sendRequest(*m, *to)
}
func (network *Network) SendFindContactMessage(to *Contact, target KademliaID) {
	m := new(msg)
	m.Method = FindNode
	m.Caller = network.msgHeader
	m.Payload.FindNode = target
	network.outRequest.mutex.Lock()
	network.outRequest.outgoingRegister[*to.ID] += 1
	network.outRequest.mutex.Unlock()
	network.sendRequest(*m, *to)
}

// Return wrapper of contact lists or data.
// func (network *Network) SendFindDataMessage(hash string) {
// 	//TODO
// }

// Return nothing as we're simply passing data to others to handle
// func (network *Network) SendStoreMessage(data []byte) {
// 	// TODO
// }
