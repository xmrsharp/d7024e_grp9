package d7024e

import (
<<<<<<< HEAD
	"fmt"
	"log"
	"strconv"
=======
	"D7024E_GRP9/kademlia/datastore"
>>>>>>> Added datastore + testing for datastore
)

// Basically the main struct of cademlia, this is the ocntainer whihc holds all the other containers.
type Kademlia struct {
<<<<<<< HEAD
	server            *Network
	ch_network_input  <-chan msg
	ch_network_output chan<- msg
	ch_node_lookup    chan []Contact
	routingTable      *RoutingTable
}

// NOTE Addrs in both server & contact.
// TODO Revisit the channel types to send messages.
func InitKademlia(ip string, port int) *Kademlia {
	ch_network_input := make(chan msg)
	ch_network_output := make(chan msg)
	ch_node_lookup := make(chan []Contact)
	addrs := ip + ":" + strconv.Itoa(port)
	server := InitNetwork(addrs, ch_network_input, ch_network_output)
	selfContact := NewContact(NewRandomKademliaID(), addrs)
	routingTable := NewRoutingTable(selfContact)
	kademlia_node := Kademlia{server, ch_network_input, ch_network_output, ch_node_lookup, routingTable}
	return &kademlia_node
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	// 1: Check if contact exists withing routing table.
	// 2: If it exists then return the contact.
	// 3: else get the closest candidates to the node
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.server.respondFindContactMessage(&node.routingTable.me, caller, candidates)
}

// AKA Node lookup
// TODO Check if i need to clear channel before each call to LUC
// TODO Change hardcoded K value (set in routing table or pass to routing etc.)
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
				// Already consumed contact - dead end (false positive).
				active_calls--
			} else {
				consumed_candidates[temp_contact.ID] = 1
				go node.server.SendFindContactMessage(&node.routingTable.me, &temp_contact, *target)
			}
			active_calls++
		}
		// for every read of new candidates, check if closer to target. if not
		// Increase the number of visited nodes w/o finding closer contact.
		temp_new_candidates = <-node.ch_node_lookup
		log.Println("FIRST NODE SHOULD HAVE SAME DISTANCE SET", temp_new_candidates[0].distance)
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
		current_candidates.Sort()
		log.Println("SIZE AFTER APPEND:", current_candidates.Len())
		active_calls--
	}
	return current_closest_node
=======
	test datastore.DataStore
}

func(node *Kademlia) Init()

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
>>>>>>> Added datastore + testing for datastore
}

// Check if target is within candidates, if not resend call to candidates of target.
func (node *Kademlia) FindContact(target Contact, candidates []Contact) {
	// CHECK IF CANDIDATES CONTAINS TARGET -> RETURN TARGET.
	for i := 0; i < len(candidates); i++ {
		//TODO Ping candidates.
		// Create go routine, if in time get response from pong -> (maybe create some local datastructure list of queues etc.)
		// If pong -> add to routing table & send request to findContact.
		//TODO Shieet, need to check if target is to be found...
	}

}

func (node *Kademlia) LookupData(hash string) {
	// TODO
}

<<<<<<< HEAD
func (node *Kademlia) Store(data []byte) {
	// TODO
=======
func (node *Kademlia) Store(data *string) {
	node.test.Insert(*data)
>>>>>>> Added datastore + testing for datastore
}

func (node *Kademlia) genCheckBuckets() {
	for i := 0; i < 20; i++ {
		num_entries := node.routingTable.buckets[i].Len()
		fmt.Println("BUCKET:", i, " Contains ", num_entries, " number of entries.")
		for j := node.routingTable.buckets[i].list.Front(); j != nil; j = j.Next() {
			fmt.Println("V:", j.Value)
		}
	}

}

func (node *Kademlia) Run() {
	log.Println("<<		STARTING NODE	>>")
	go node.server.Listen()
	for {
		// Routing table debugging
		//fmt.Println("PRE ROUTING TABLE");
		//node.genCheckBuckets()
		server_msg := <-node.ch_network_input
		log.Println("RECIEVED: ", server_msg.Caller, "FROM CH")
		if server_msg.Caller.Address == node.routingTable.me.Address {
			log.Println("SUSPECT CALLER: IDENTICAL ADDRS OFF:", node.routingTable.me.Address)
		} else {
			node.routingTable.AddContact(server_msg.Caller)
			// TODO In all requests check if candidates are not nil -> means we need to resend to candidates the same request.
			switch server_msg.Method {
			case Ping:
				if server_msg.Payload.PingPong == "PING" {
					log.Println("RECIEVED PING EVENT")
					node.server.SendPingMessage(&node.routingTable.me, &server_msg.Caller, "PONG")
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
					// NODE LOOKUP AND FIND NODE NOT THE SAME THING.
					// THE FINDNODE IS SIMPLY -> REQUEST TARGET RETURN K CLOSEST.
					node.ch_node_lookup <- server_msg.Payload.Candidates
					// FOR NODE LOOKUP -> INITIATE FIND NODE, IF RESPONSE CONTAINS TARGET OF
					// FIND NODE, ADD TO CHANNEL SPECIFIC FOR NODE LOOKUP INSTANCE.

					// Then simply request a of the other k closests beginning nodes.

					// Lookup terminates when it has gotten response of K closest
					// Response, containing candidates. How to keep track of target her
					// Not nil, so this is a response.
					// Meaning we need to check if target can be found within candidates.
					// And if they are. send to command channel or w/e that contact is found at some address
				}
			case FindValue:
				log.Println("RECIEVED FINDVALUE EVENT")
			default:
				log.Println("RECIEVED OTHER EVENT")
			}

		}
		//fmt.Println("POST ROUTING TABLE");
		//node.genCheckBuckets()

	}

}
