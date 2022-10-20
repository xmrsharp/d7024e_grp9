package kademlia

import (
	"log"
	"net"
)

type Network struct {
	msgHeader        Contact
	outgoingRequests *OutgoingRegister
	addrs            *net.UDPAddr
	channelWriteNode chan<- msg
}

const (
	PACKET_BYTE_SIZE = 8192
)

func NewNetwork(msgHeader Contact, addrs string, outgoingRequests *OutgoingRegister, write chan<- msg) *Network {
	udpAddr, err := net.ResolveUDPAddr("udp", addrs)
	if err != nil {
		log.Panicf(("CANNOT SERVE ON SPEC ADDR - %s"), err)
	}
	network := Network{msgHeader, outgoingRequests, udpAddr, write}

	return &network
}

func (network *Network) Listen() {
	serverSocket, err := net.ListenUDP("udp", network.addrs)
	if err != nil {
		log.Println(err)
	}
	log.Println("KAD NODE SERVING ON:", network.addrs)
	defer serverSocket.Close()
	for {
		// Change size of reader.
		buff := make([]byte, PACKET_BYTE_SIZE)
		n, caller_addr, err := serverSocket.ReadFromUDP(buff)
		if err != nil {
			log.Println("FAILED TO READ SOCKET:", err)
		} else {
			go network.handleRequest(buff[:n], caller_addr)
		}
	}
}

func (network *Network) handleRequest(m []byte, addr *net.UDPAddr) {
	decodedMsg, err := decodeMsg(m)
	if err != nil {
		log.Println("UNKNOWN MSG BY:", addr)
		return
	}
	network.channelWriteNode <- decodedMsg
}

func (network *Network) sendRequest(m msg, to Contact) {
	if m.Method != 0 {
		log.Printf(("[%s] SENDING [%s] TO: [%s]"), network.addrs, m.Method, to.Address)
	}
	payload, err := encodeMsg(m)
	if err != nil {
		log.Printf(("ENCODING ERROR: %s"), err)
	}
	udpAddr, err := net.ResolveUDPAddr("udp", to.Address)
	if err != nil {
		log.Printf(("RESOLVE ADDR ERR TO [%s] - %s"), to.Address, err)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()
	if err != nil {
		log.Printf(("FAILED TO ESTABLISH SOCKET TO [%s] - %s"), to.Address, err)
	}
	_, err = conn.Write(payload)
	if err != nil {
		log.Printf(("FAILED TO WRITE - %s"), err)
	}
}

func (network *Network) respondPingMessage(to *Contact) {
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

func (network *Network) respondFindDataMessage(to *Contact, data string) {
	m := new(msg)
	m.Method = FindValue
	m.Caller = network.msgHeader
	m.Payload.FindValue.Value = data
	network.sendRequest(*m, *to)
}

func (network *Network) SendFindDataMessage(to *Contact, key KademliaID) {
	m := new(msg)
	m.Method = FindValue
	m.Caller = network.msgHeader
	m.Payload.FindValue.Key = key
	network.sendRequest(*m, *to)
}

func (network *Network) respondStoreMessage(to *Contact, key KademliaID) {
	m := new(msg)
	m.Method = Store
	m.Caller = network.msgHeader
	m.Payload.Store.Key = key
	network.sendRequest(*m, *to)
}

func (network *Network) SendStoreMessage(to *Contact, target KademliaID, data string) {
	m := new(msg)
	m.Method = Store
	m.Caller = network.msgHeader
	m.Payload.Store.Key = target
	m.Payload.Store.Value = data
	network.sendRequest(*m, *to)
}
