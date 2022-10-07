package d7024e

import (
	"log"
	"net"
	"sync"
)

type Network struct {
	msgHeader        Contact
	wg               sync.WaitGroup
	outgoingRequests *SharedMap
	addrs            *net.UDPAddr
	channelWriteNode chan<- msg
	channelReadNode  <-chan msg
}

func NewNetwork(msgHeader Contact, addrs string, outgoingRequests *SharedMap, write chan<- msg, read <-chan msg) *Network {
	udpAddr, err := net.ResolveUDPAddr("udp4", addrs)
	if err != nil {
		log.Panic("CANNOT SERVE ON SPECIFIED ADDR")
	}
	network := Network{msgHeader, sync.WaitGroup{}, outgoingRequests, udpAddr, write, read}
	return &network
}

func (network *Network) Listen() {
	serverSocket, err := net.ListenUDP("udp4", network.addrs)
	if err != nil {
		log.Println(err)
	}
	log.Println("SERVING ON:", network.addrs)
	defer serverSocket.Close()
	for {
		// Change size of reader.
		buff := make([]byte, 4096, 4096)
		n, caller_addr, err := serverSocket.ReadFromUDP(buff)
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
	decodedMsg, err := decodeMsg(m)
	if err != nil {
		// Simply want to end routine nicely.
		log.Println("UNKNOWN MSG BY:", addr)
		return
	}
	network.channelWriteNode <- decodedMsg
}

// Add caller to awaiting result channel?
// Here need to alter the dialup to simply call the addrs from contact.
func (network *Network) sendRequest(m msg, to Contact) {
	payload, _ := encodeMsg(m)
	udpAddr, err := net.ResolveUDPAddr("udp4", to.Address)
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO ESTABLISH CONNECTION TO: ", to)
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO ESTABLISH CONNECTION TO: ", to)
	}
	_, err = conn.Write(payload)
	defer conn.Close()
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Println("FAILED TO WRITE TO: ", udpAddr)
	} else {
		log.Println("SENT REQ:", m.Method, " TO: ", to.Address)
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
	network.outgoingRequests.mutex.Lock()
	network.outgoingRequests.outgoingRegister[*to.ID] += 1
	network.outgoingRequests.mutex.Unlock()
	network.sendRequest(*m, *to)
}
func (network *Network) SendFindContactMessage(to *Contact, target KademliaID) {
	m := new(msg)
	m.Method = FindNode
	m.Caller = network.msgHeader
	m.Payload.FindNode = target
	network.outgoingRequests.mutex.Lock()
	network.outgoingRequests.outgoingRegister[*to.ID] += 1
	network.outgoingRequests.mutex.Unlock()
	network.sendRequest(*m, *to)
}

// Return wrapper of contact lists or data.
func (network *Network) SendFindDataMessage(to *Contact, key KademliaID, data [20]byte) {
	m := new(msg)
	m.Method = FindValue
	m.Caller = network.msgHeader
	m.Payload.Value = data
	network.sendRequest(*m, *to)
}

//Return nothing as we're simply passing data to others to handle
func (network *Network) SendStoreMessage(to *Contact, target KademliaID, data [20]byte, status chan bool) {
	//
	m := new(msg)
	m.Method = Store
	m.Caller = network.msgHeader
	m.Payload.Store.Key = target
	m.Payload.Store.Value = data
	network.outgoingRequests.mutex.Lock()
	network.outgoingRequests.outgoingRegister[*to.ID] += 1
	network.outgoingRequests.mutex.Unlock()
	network.sendRequest(*m, *to)
	status <- true

}
