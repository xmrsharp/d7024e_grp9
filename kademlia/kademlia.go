package d7024e

import (
	"log"
	"strconv"
)

// Basically the main struct of cademlia, this is the ocntainer whihc holds all the other containers.
type Kademlia struct {
	server            *Network
	ch_network_input  <-chan msg
	ch_network_output chan<- msg
	routingTable      *RoutingTable
}

// TODO Revisit the channel types to send messages.
func InitKademlia(ip string, port int) *Kademlia {
	ch_network_input := make(chan msg)
	ch_network_output := make(chan msg)
	addrs := ip + ":" + strconv.Itoa(port)
	// Addrs in both server & contact.
	server := InitNetwork(addrs, ch_network_input, ch_network_output)
	selfContact := NewContact(NewRandomKademliaID(), addrs)
	routingTable := NewRoutingTable(selfContact)
	kademlia_node := Kademlia{server, ch_network_input, ch_network_output, routingTable}
	return &kademlia_node
}
func (node *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (node *Kademlia) LookupData(hash string) {
	// TODO
}

func (node *Kademlia) Store(data []byte) {
	// TODO
}

func (node *Kademlia) Run() {
	log.Println("<<		STARTING NODE	>>")
	go node.server.Listen()
	for {
		server_msg := <-node.ch_network_input
		log.Println("RECIEVED: ", server_msg, "FROM CH")
		// NOTE In all incoming requests check to see if we can update the calle address to routing table.
		node.routingTable.AddContact(server_msg.Caller)
		// TODO In all requests check if candidates are not nil -> means we need to resend to candidates the same request.
		switch server_msg.Method {
		case Ping:
			log.Println("RECIEVED PING EVENT")
			ping_payload := new(content)
			ping_payload.PingPong = "PONG"
			test_msg := msg{Method: Ping, Caller: server_msg.Caller, Payload: *ping_payload}
			node.server.sendRequest(test_msg, server_msg.Caller)
		case Store:
			log.Println("RECIEVED STORE EVENT")
		case FindNode:
			log.Println("RECIEVED FINDNODE EVENT")
		case FindValue:
			log.Println("RECIEVED FINDVALUE EVENT")
		default:
			log.Println("RECIEVED OTHER EVENT")
		}
		log.Println("NEXT REQUEST INC")
	}

}
