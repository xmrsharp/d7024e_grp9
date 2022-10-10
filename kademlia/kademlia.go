package d7024e

import (
	"fmt"
	"log"
	"os"
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
	sm.mutex.Lock()
	for _, v := range sm.outgoingRegister {
		if v > 0 {
			sm.mutex.Unlock()
			return true
		}
	}
	sm.mutex.Unlock()
	return false

}

// Logic and state of Kademlia node
type Kademlia struct {
	server              *Network
	outgoingRequests    *SharedMap
	channelServerInput  <-chan msg
	channelServerOutput chan<- msg
	channelNodeLookup   chan []Contact
	routingTable        *RoutingTable
	datastore           DataStore
}

func NewKademlia(ip string, port int, id *KademliaID) *Kademlia {
	channelServerInput := make(chan msg)
	channelServerOutput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	addrs := ip + ":" + strconv.Itoa(port)
	outgoingRequests := NewSharedMap()
	selfContact := NewContact(id, addrs)
	server := NewNetwork(selfContact, addrs, outgoingRequests, channelServerInput, channelServerOutput)
	routingTable := NewRoutingTable(selfContact)
	datastore := NewDataStore()
	kademliaNode := Kademlia{server, outgoingRequests, channelServerInput, channelServerOutput, channelNodeLookup, routingTable, datastore}

	return &kademliaNode
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.server.respondFindContactMessage(caller, candidates)
}

func (node *Kademlia) NodeLookup(target *KademliaID) {
	// Initiate candidates.
	currentCandidates := ContactCandidates{node.routingTable.FindClosestContacts(target, BucketSize)}

	// Inititate end condition.
	currentClosestNode := currentCandidates.GetClosest()
	probedNoCloser := 0

	consumedCandidates := make(map[*KademliaID]int)

	// For keeping track off active outgoing FindNode RPCs.
	activeAlphaCalls := 0

	var newCandidates []Contact
	for probedNoCloser < BucketSize {
		if currentClosestNode.ID.Equals(target) {
			return
		}
		for activeAlphaCalls < ALPHA_VALUE && currentCandidates.Len() > 0 {
			tempContact := currentCandidates.PopClosest()
			// Check for alrdy consumed nodes.
			if consumedCandidates[tempContact.ID] == 1 {
				// Already consumed contact - dead end.
				activeAlphaCalls--
			} else {
				consumedCandidates[tempContact.ID] = 1
				node.server.SendFindContactMessage(&tempContact, *target)
			}
			activeAlphaCalls++
		}
		if !node.outgoingRequests.expectingIncRequest() {
			return
		}
		newCandidates = <-node.channelNodeLookup
		// Investigate to why I have to recalculate the distances?
		for i := 0; i < len(newCandidates); i++ {
			newCandidates[i].CalcDistance(target)
			node.server.SendPingMessage(&(newCandidates[i]))
		}

		// Only need to check head of list as list is ordered on arrival.
		if !currentClosestNode.Less(&newCandidates[0]) {
			// Got closer to target, update current closest and succesfull probes.
			probedNoCloser = 0
			currentClosestNode = newCandidates[0]
		} else {
			// No closer to target.
			probedNoCloser++
		}
		currentCandidates.Append(newCandidates)
		activeAlphaCalls--
	}
}

func (node *Kademlia) FindClosestContacts(target *KademliaID, count int) {
	node.NodeLookup(target)
}

func (node *Kademlia) LookupData(hash string) (string, string) {
	// TODO
	return hash, node.routingTable.me.Address //ändra så att den returnar rätt värde och IP addressen eller node namnet det var i
}

func (node *Kademlia) Store(data *string) string {
	node.datastore.Insert(data)
	return node.datastore.Get(*node.routingTable.me.ID)
}

func (node *Kademlia) bootLoader(bootLoaderAddrs string, bootLoaderID KademliaID) {
	if bootLoaderAddrs == "" {
		return
	}
	bootContact := NewContact(&bootLoaderID, bootLoaderAddrs)
	node.routingTable.AddContact(bootContact)
	node.NodeLookup(node.routingTable.me.ID)
}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	go node.server.Listen()
	go node.bootLoader(bootLoaderAddrs, bootLoaderID)
	go Cli(os.Stdout, node)
	for {
		serverMsg := <-node.channelServerInput
		node.routingTable.AddContact(serverMsg.Caller)
		log.Printf(("RECIEVED [%s] EVENT FROM [%s]:[%s]"), serverMsg.Method.String(), serverMsg.Caller.Address, serverMsg.Caller.ID)
		if serverMsg.Caller.Address == node.routingTable.me.Address {
			log.Printf("SUSPECT CALLER [%s] REASON: IDENTICAL ADDRS AS SERVER", serverMsg.Caller.ID)
		} else {
			switch serverMsg.Method {
			case Ping:
				if serverMsg.Payload.PingPong == "PING" {
					node.server.SendPongMessage(&serverMsg.Caller)
				}
			case Store:
				// TODO Handle inc store event.
			case FindNode:
				if serverMsg.Payload.Candidates == nil {
					node.routingTable.AddContact(serverMsg.Caller)
					node.ReturnCandidates(&serverMsg.Caller, &serverMsg.Payload.FindNode)
				} else {
					node.outgoingRequests.mutex.Lock()
					if node.outgoingRequests.outgoingRegister[*serverMsg.Caller.ID] > 0 {
						node.routingTable.AddContact(serverMsg.Caller)
						node.outgoingRequests.outgoingRegister[*serverMsg.Caller.ID] -= 1
						node.outgoingRequests.mutex.Unlock()
						node.channelNodeLookup <- serverMsg.Payload.Candidates
					} else {
						log.Printf("SUSPECT CALLER [%s] REASON: UNEXPECTED FIND_NODE RESPONSE", serverMsg.Caller.ID)
						node.outgoingRequests.mutex.Unlock()
					}
				}
			case FindValue:
				// TODO Handle inc find value event.
			default:
				log.Println("PANIC - UNKNOWN RPC METHOD")
			}

		}
	}
}

// NOTE ONLY TO BE USED FOR TESTING/DEMONSTRATION PURPOSES ONLY
func (node *Kademlia) genCheckBuckets() {
	count := 0
	for i := 0; i < 20; i++ {
		//num_entries := node.routingTable.buckets[i].Len()
		// fmt.Println("BUCKET:", i, " Contains ", num_entries, " number of entries.")
		for j := node.routingTable.buckets[i].list.Front(); j != nil; j = j.Next() {
			// NOTE If individual number of entries are to be inspected.
			//fmt.Println("V:", j.Value)
			count++
		}
	}
	fmt.Println("TOTAL NUMBER OF UNIQUE ENTRIES : ", count)
}
