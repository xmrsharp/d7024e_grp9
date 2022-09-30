package d7024e

import (
	"fmt"
	"log"
	"strconv"
	"sync"
)

const (
	ALPHA_VALUE = 3
)

// SharedMap used by Kademlia and network to synchronize expected incoming responses by other nodes.
type SharedMap struct {
	mutex            *sync.Mutex
	outgoingRegister map[KademliaID]int
}

func NewSharedMap() *SharedMap {
	return &SharedMap{&sync.Mutex{}, make(map[KademliaID]int)}
}

func (sm *SharedMap) expectingIncRequest() bool {
	sm.mutex.Lock
	for k, v := range sm.outgoingRegister {
		if v > 1 {
			log.Println("--> Expecting response from:", k)
			return true
		}
	}
	return false

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
	datastore         DataStore
}

// TODO Go over variable names, they're currently trash
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
	datastore := NewDataStore()
	kademlia_node := Kademlia{server, outRequest, ch_network_input, ch_network_output, ch_node_lookup, ch_node_response, routingTable, datastore}
	return &kademlia_node
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.server.respondFindContactMessage(caller, candidates)
}

func (node *Kademlia) NodeLookup(target *KademliaID) Contact {
	// Initiate candidates from own routing table.
	current_candidates := ContactCandidates{node.routingTable.FindClosestContacts(target, BucketSize)}

	// Inititate end condition.
	current_closest_node := current_candidates.GetClosest()
	probed_no_closer := 0

	consumed_candidates := make(map[*KademliaID]int)

	// For keeping track off outgoing FindNode RPCs.
	active_calls := 0

	var temp_new_candidates []Contact

	for probed_no_closer < BucketSize && node.outRequest.expectingIncRequest() {
		log.Println("START OF OUTER LOOP")
		if current_closest_node.ID.Equals(target) {
			return // node.routingTable.FindClosestContacts(target, BucketSize)
		}
		for active_calls < ALPHA_VALUE && current_candidates.Len() > 0 {
			temp_contact := current_candidates.PopClosest()
			// Check for alrdy consumed nodes
			if consumed_candidates[temp_contact.ID] == 1 {
				// Already consumed contact - dead end.
				active_calls--
			} else {
				consumed_candidates[temp_contact.ID] = 1
				node.server.SendFindContactMessage(&temp_contact, *target)
			}
			active_calls++
		}
		// Increase the number of visited nodes w/o finding closer contact.
		temp_new_candidates = <-node.ch_node_lookup
		// Investigate to why I have to recalculate the distances?
		for i := 0; i < len(temp_new_candidates); i++ {
			temp_new_candidates[i].CalcDistance(target)
			go node.server.SendPingMessage(temp_new_candidates[i], "PING")
		}

		// Only need to check head of list as list is ordered on arrival.
		if !current_closest_node.Less(&temp_new_candidates[0]) {
			// Got closer to target, update current closest and succesfull probes.
			probed_no_closer = 0
			current_closest_node = temp_new_candidates[0]
		} else {
			// No closer to target.
			probed_no_closer++
		}
		current_candidates.Append(temp_new_candidates)
		// TODO REMOVE: ONLY FOR TESTING PURPOSES
		//for i := 0; i < current_candidates.Len(); i++ {
		//	log.Println(current_candidates.contacts[i].distance)
		//	log.Println("CANDIDATE", i, ": ", current_candidates.contacts[i].String())
		//}

		active_calls--
	}
	// return node.routingTable.FindClosestContacts(target, BucketSize)
}

func (node *Kademlia) FindClosestContacts(target *KademliaID, count int) []Contact {
	node.NodeLookup(target)
	return node.routingTable.FindClosestContacts(target, count)
}

func (node *Kademlia) LookupData(hash string) {
	// TODO
}

func (node *Kademlia) Store(data *string) {
	node.datastore.Insert(*data)
}

func (node *Kademlia) bootLoader(bootLoaderAddrs string, bootLoaderID KademliaID) {
	if bootLoaderAddrs == "" {
		return
	}
	bootContact := NewContact(&bootLoaderID, bootLoaderAddrs)
	node.routingTable.AddContact(bootContact)
	// Run node lookup to populate routing table.
	node.NodeLookup(node.routingTable.me.ID)
	log.Println("Closest neighbours:", node.routingTable.FindClosestContacts(node.routingTable.me.ID, BucketSize))
}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	log.Println("<<		STARTING NODE	>>")
	go node.server.Listen()
	go node.bootLoader(bootLoaderAddrs, bootLoaderID)
	for {
		server_msg := <-node.ch_network_input
		log.Println("MAIN LOOP RECIEVED: ", server_msg.Caller, "FROM CH")
		if server_msg.Caller.Address == node.routingTable.me.Address {
			log.Println("SUSPECT CALLER: IDENTICAL ADDRS OFF:", node.routingTable.me.Address)
		} else {
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
				log.Println("RECIEVED FINDNODE EVENT")
				if server_msg.Payload.Candidates == nil {
					// Inc request, so simply return candidates.
					node.ReturnCandidates(&server_msg.Caller, &server_msg.Payload.FindNode)
				} else {
					node.outRequest.mutex.Lock()
					if node.outRequest.outgoingRegister[*server_msg.Caller.ID] > 0 {
						node.outRequest.outgoingRegister[*server_msg.Caller.ID] -= 1 // node.outRequest.outgoingRegister[*server_msg.Caller.ID] - 1
						node.outRequest.mutex.Unlock()
						node.ch_node_lookup <- server_msg.Payload.Candidates
					} else {
						// Possible malicious request.
						node.outRequest.mutex.Unlock()
					}
				}
			case FindValue:
				log.Println("RECIEVED FINDVALUE EVENT")
			default:
				log.Println("RECIEVED OTHER EVENT")
			}

		}
		// Will add any incoming request caller to routingTable if possible.
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
