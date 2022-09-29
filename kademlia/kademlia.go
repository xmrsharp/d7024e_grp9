package d7024e

import (
	"fmt"
	"log"
	"strconv"
	"sync"
)

type SharedMap struct {
	mutex            *sync.Mutex
	outgoingRegister map[KademliaID]int
}

// SharedMap used by Kademlia and network to synchronize expected incoming responses by other nodes.
func NewSharedMap() *SharedMap {
	return &SharedMap{&sync.Mutex{}, make(map[KademliaID]int)}
}

// Logic and state of Kademlia node
type Kademlia struct {
	server            *Network
	outRequest        *SharedMap
	ch_network_input  <-chan msg
	ch_network_output chan<- msg
	ch_node_lookup    chan []Contact
	ch_node_response  chan []Contact
	routingTable      *RoutingTable
}

// TODO Go over variable names, they're currently trash
// NOTE Addrs in both server & contact.
func NewKademlia(ip string, port int) *Kademlia {
	ch_network_input := make(chan msg)
	ch_network_output := make(chan msg)
	ch_node_lookup := make(chan []Contact)
	ch_node_response := make(chan []Contact)
	addrs := ip + ":" + strconv.Itoa(port)
	outRequest := NewSharedMap()
	selfContact := NewContact(NewRandomKademliaID(), addrs)
	server := InitNetwork(selfContact, addrs, outRequest, ch_network_input, ch_network_output)
	routingTable := NewRoutingTable(selfContact)
	kademlia_node := Kademlia{server, outRequest, ch_network_input, ch_network_output, ch_node_lookup, ch_node_response, routingTable}
	return &kademlia_node
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.server.respondFindContactMessage(caller, candidates)
	for i := 0; i < len(candidates); i++ {
		log.Println("	SENDING:", candidates[i].String())
		log.Println(candidates[i].distance)
	}
}

// AKA Node lookup
// TODO Change hardcoded K value and alpha value (set in routing table or pass to routing etc.)
// TODO Need way of sorting to get closest nodes?
func (node *Kademlia) LookupContact(target *KademliaID) Contact {
	hard_coded_k_value := 20

	// 1. Fetch closest nodes to target within own routing table (alpha closest).
	current_candidates := ContactCandidates{node.routingTable.FindClosestContacts(target, hard_coded_k_value)}
	// Initiating end condition.
	current_closest_node := current_candidates.GetClosest()
	probed_no_closer := 0

	// Used for keeping track of consumed candidates.
	consumed_candidates := make(map[*KademliaID]int)

	// For keeping track of active requests.
	active_calls := 0

	var temp_new_candidates []Contact

	for probed_no_closer < hard_coded_k_value {
		log.Println("START OF OUTER LOOP")
		if current_closest_node.ID.Equals(target) {
			return current_closest_node
		}
		// send out request for the alpha number of connections.
		for active_calls < 3 && current_candidates.Len() > 0 {
			log.Println("START OF INNER LOOP")
			log.Println("LENGTH OF PRE CURRENT_CANDIDATES:", current_candidates.Len())
			temp_contact := current_candidates.PopClosest()
			log.Println("LENGTH OF POST CURRENT_CANDIDATES:", current_candidates.Len())
			// Check for alrdy consumed nodes
			if consumed_candidates[temp_contact.ID] == 1 {
				// Already consumed contact - dead end.
				active_calls--
			} else {
				consumed_candidates[temp_contact.ID] = 1
				go node.server.SendFindContactMessage(&temp_contact, *target)
			}
			active_calls++
		}
		// for every read of new candidates, check if closer to target. if not
		// Increase the number of visited nodes w/o finding closer contact.
		temp_new_candidates = <-node.ch_node_lookup
		// Investigate to why i have to recalculate the distances?
		// FOR EACH INCOMING NODE, SEND PING -> IF PONG (ADD TO ROUTING TABLE.)
		for i := 0; i < len(temp_new_candidates); i++ {
			temp_new_candidates[i].CalcDistance(target)
		}

		log.Println("FIRST NODE SHOULD HAVE SAME DISTANCE SET", temp_new_candidates[0].distance)
		log.Println(temp_new_candidates[0].String())
		temp_new_candidates[0].CalcDistance(target)
		log.Println("COMPARED TO", temp_new_candidates[0].distance)
		log.Println("IF SAME -> REMOVE CALC DISTANCE ABOVE")
		// Here only need to check if the head of list is closer to target as the list will be ordered on arrival (closest at head).
		if !current_closest_node.Less(&temp_new_candidates[0]) {
			// Got closer to target, update current closest and succesfull probes.
			probed_no_closer = 0
			current_closest_node = temp_new_candidates[0]
		} else {
			// No closer to target.
			probed_no_closer++
		}
		log.Println("SIZE CANDIDATES:", current_candidates.Len(), " SIZE NEW CAND:", len(temp_new_candidates))
		current_candidates.Append(temp_new_candidates)
		log.Println("DID IT WORK")
		for i := 0; i < current_candidates.Len(); i++ {
			log.Println(current_candidates.contacts[i].distance)
			log.Println("CANDIDATE", i, ": ", current_candidates.contacts[i].String())
		}
		log.Println("SIZE POST APPEND:", current_candidates.Len())

		//(&current_candidates).Sort()
		log.Println("SIZE AFTER APPEND:", current_candidates.Len())
		active_calls--
	}
	return current_closest_node
}

func (node *Kademlia) LookupData(hash string) {
	// TODO
}

func (node *Kademlia) Store(data []byte) {
	// TODO
}

func (node *Kademlia) bootLoader() {
	// Simply perform node lookup on self to supplied main node.
	// Hard coded for now for testing purpouses.
	addrs := "192.168.38.201:8888"
	boot_loader_id := NewKademliaID("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	boot_contact := NewContact(boot_loader_id, addrs)
	// Append contact to routing table and search for self.
	node.routingTable.AddContact(boot_contact)
	log.Println("BOOT RETURN VALUE: ", node.LookupContact(node.routingTable.me.ID))
}

func (node *Kademlia) Run() {
	log.Println("<<		STARTING NODE	>>")
	go node.server.Listen()
	go node.bootLoader()
	for {
		// Routing table debugging
		//fmt.Println("PRE ROUTING TABLE");
		//node.genCheckBuckets()
		server_msg := <-node.ch_network_input
		log.Println("MAIN LOOP RECIEVED: ", server_msg.Caller, "FROM CH")
		if server_msg.Caller.Address == node.routingTable.me.Address {
			log.Println("SUSPECT CALLER: IDENTICAL ADDRS OFF:", node.routingTable.me.Address)
		} else {
			log.Println("CHECKING METHOD")
			//node.routingTable.AddContact(server_msg.Caller)
			switch server_msg.Method {
			case Ping:
				if server_msg.Payload.PingPong == "PING" {
					log.Println("RECIEVED PING EVENT")
					node.server.SendPingMessage(&server_msg.Caller, "PONG")
				} else {
					log.Println("RECIEVED PONG EVENT")
				}
			case Store:
				log.Println("RECIEVED STORE EVENT")
			case FindNode:
				if server_msg.Payload.Candidates == nil {
					log.Println("RECIEVED FINDNODE EVENT -> WITHOUT CANDIDATES PAYLOAD")
					// Inc request, so simply return candidates.
					node.ReturnCandidates(&server_msg.Caller, &server_msg.Payload.FindNode)
				} else {
					log.Println("RECIEVED FINDNODE EVENT -> WITH CANDIDATES PAYLOAD")
					test_candidate := server_msg.Payload.Candidates
					for i := 0; i < len(test_candidate); i++ {
						log.Println("RECIEVED:", test_candidate[i].String())

					}
					node.outRequest.mutex.Lock()
					if node.outRequest.outgoingRegister[*server_msg.Caller.ID] > 0 {
						node.outRequest.outgoingRegister[*server_msg.Caller.ID] = node.outRequest.outgoingRegister[*server_msg.Caller.ID] - 1
						node.outRequest.mutex.Unlock()
						node.ch_node_lookup <- server_msg.Payload.Candidates
					} else {
						// Simply want to insert into channel if requested.
						node.outRequest.mutex.Unlock()
					}
				}
			case FindValue:
				log.Println("RECIEVED FINDVALUE EVENT")
			default:
				log.Println("RECIEVED OTHER EVENT")
			}

		}
		node.routingTable.AddContact(server_msg.Caller)
	}

}

// TODO REMOVE: USED FOR TESTING PURPOSES ONLY
func (node *Kademlia) genCheckBuckets() {
	for i := 0; i < 20; i++ {
		num_entries := node.routingTable.buckets[i].Len()
		fmt.Println("BUCKET:", i, " Contains ", num_entries, " number of entries.")
		for j := node.routingTable.buckets[i].list.Front(); j != nil; j = j.Next() {
			fmt.Println("V:", j.Value)
		}
	}

}
