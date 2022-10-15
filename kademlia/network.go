package kademlia

import (
	"log"
	"net"
	"sync"
)

type Network struct {
	msgHeader        Contact
	wg               sync.WaitGroup
	outgoingRequests *OutgoingRegister
	addrs            *net.UDPAddr
	channelWriteNode chan<- msg
	channelReadNode  <-chan msg
}

func NewNetwork(msgHeader Contact, addrs string, outgoingRequests *OutgoingRegister, write chan<- msg, read <-chan msg) *Network {
	udpAddr, err := net.ResolveUDPAddr("udp", addrs)
	if err != nil {
		log.Panicf(("CANNOT SERVE ON SPEC ADDR - %s"), err)
	}
	network := Network{msgHeader, sync.WaitGroup{}, outgoingRequests, udpAddr, write, read}
	return &network
}

func (network *Network) Listen() {
	serverSocket, err := net.ListenUDP("udp", network.addrs)
	if err != nil {
		log.Println(err)
	}
	log.Println("SERVING ON:", network.addrs)
	defer serverSocket.Close()
	for {
		// Change size of reader.
		buff := make([]byte, 8192)
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
	log.Printf(("SENDING [%s] TO: [%s]"), m.Method, to.Address)
	payload, err := encodeMsg(m)
	if err != nil {
		log.Printf(("ENCODING ERROR: %s"), err)
	}
	udpAddr, err := net.ResolveUDPAddr("udp", to.Address)
	if err != nil {
		log.Printf(("RESOLVE ADDR ERR TO [%s] - %s"), to.Address, err)
		// TODO Here alert node to remove contact from routing table
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		// TODO Here alert node to remove contact from routing table
		log.Printf(("FAILED TO ESTABLISH SOCKET TO [%s] - %s"), to.Address, err)
	}
	_, err = conn.Write(payload)
	if err != nil {
		log.Printf(("FAILED TO WRITE - %s"), err)
	}
}

func (network *Network) SendPongMessage(to *Contact) {
	m := new(msg)
	m.Method = Ping
	m.Caller = network.msgHeader
	m.Payload.PingPong = "PONG"
	network.sendRequest(*m, *to)
}
func (network *Network) SendPingMessage(to *Contact) {
	m := new(msg)
	m.Method = Ping
	m.Caller = network.msgHeader
	m.Payload.PingPong = "PING"
	network.sendRequest(*m, *to)
}

func (network *Network) respondFindContactMessage(to *Contact, candidates []Contact) {
	m := new(msg)
	m.Method = FindNode
	m.Caller = network.msgHeader
	m.Payload.Candidates = candidates
	network.sendRequest(*m, *to)
}
func (network *Network) SendFindContactMessage(to *Contact, target KademliaID) {
	m := new(msg)
	m.Method = FindNode
	m.Caller = network.msgHeader
	m.Payload.FindNode = target
	network.outgoingRequests.RegisterOutgoing(*to.ID)
	network.sendRequest(*m, *to)
}

// Return wrapper of contact lists or data.
func (network *Network) SendFindDataMessage(to *Contact, key KademliaID) {
	m := new(msg)
	m.Method = FindValue
	m.Caller = network.msgHeader
	m.Payload.Key = key
	m.Payload.Store.Key = key
	log.Printf("SendFindDataMessage called with KEY: %s", key.String())
	network.sendRequest(*m, *to)
}
func (network *Network) SendReturnDataMessage(to *Contact, data string) {
	m := new(msg)
	m.Method = FindValue
	m.Caller = network.msgHeader
	m.Payload.Value = data
	network.sendRequest(*m, *to)
}

// Return nothing as we're simply passing data to others to handle
func (network *Network) SendStoreMessage(to *Contact, target KademliaID, data string, status chan bool) {
	//
	m := new(msg)
	m.Method = Store
	m.Caller = network.msgHeader
	m.Payload.Store.Key = target
	m.Payload.Store.Value = data
	network.sendRequest(*m, *to)
	status <- true

}
