package d7024e

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	ALPHA_VALUE = 3
	K_VALUE     = 20
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

// TODO Go over variable names, they're currently trash
func NewKademlia(ip string, port int) *Kademlia {
	channelServerInput := make(chan msg)
	channelServerOutput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	addrs := ip + ":" + strconv.Itoa(port)
	outgoingRequests := NewSharedMap()
	selfContact := NewContact(NewRandomKademliaID(), addrs)
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
			// Not expecting any response - entered loop with only consumed candidates in current candidates - return to avoid block.
			return
		}

		newCandidates = <-node.channelNodeLookup
		// Investigate to why I have to recalculate the distances?
		for i := 0; i < len(newCandidates); i++ {
			newCandidates[i].CalcDistance(target)
			go node.server.SendPingMessage(&(newCandidates[i]), "PING")
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

func (node *Kademlia) FindClosestContacts(target *KademliaID, count int) []Contact {
	node.NodeLookup(target)
	return node.routingTable.FindClosestContacts(target, count)
}

/*
 	Check that the hash is valid
 	Check if the data is found on this node
	Send request to look for the data on other nodes
		NodeLookup (key = target)
		Send request to all contacts found in the lookup.
*/
func (node *Kademlia) LookupData(hash string) {
	// TODO
}

/*
Create new random ID
Update routing table for the new id
Get closest contacts for id and try sending store message to them
*/

func (node *Kademlia) Store(data []byte, status chan bool) {
	log.Println("Store command in kademlia.go called with data:")
	log.Println(data)
	hashed := Hash(data)
	//str2B := []byte(hashed)
	key := NewKademliaID(&hashed)

	//contacts := kademlia.LookupContact((node.KademliaID)(str2B))
	log.Println("Trying to store key: " + key.String())
	node.NodeLookup(&key)
	//Array with contacs
	neighbours := node.FindClosestContacts(&key, K_VALUE)
	//Loop through closest contacts and try storing on them.
	for i := 0; i < len(neighbours); i++ {
		storeStatus := make(chan bool)
		go node.server.SendStoreMessage(&(neighbours[i]), key, data, storeStatus)
		select {
		case <-storeStatus:
			log.Println("Stored data on node: " + (neighbours[i]).ID.String())
		case <-time.After(10 * time.Second):
			log.Println("TOOK TOO LONG TIME!!")
		}
	}
	status <- true

}

func Hash(data []byte) string {
	sha1 := sha1.Sum([]byte(data))
	key := hex.EncodeToString(sha1[:])

	return key
}

func (node *Kademlia) bootLoader(bootLoaderAddrs string, bootLoaderID KademliaID) {
	if bootLoaderAddrs == "" {
		return
	}
	bootContact := NewContact(&bootLoaderID, bootLoaderAddrs)
	node.routingTable.AddContact(bootContact)

}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	log.Println("<<		STARTING NODE	>>")
	go node.server.Listen()
	go node.bootLoader(bootLoaderAddrs, bootLoaderID)
	for {
		serverMsg := <-node.channelServerInput
		log.Printf("RECIEVED %s EVENT FROM [%s]:%s", serverMsg.Method.String(), serverMsg.Caller.ID, serverMsg.Caller.Address)
		node.routingTable.AddContact(serverMsg.Caller)
		if serverMsg.Caller.Address == node.routingTable.me.Address {
			log.Printf("SUSPECT CALLER [%s] REASON: IDENTICAL ADDRS AS SERVER", serverMsg.Caller.ID)
		} else {
			switch serverMsg.Method {
			case Ping:
				log.Println("PING")
				if serverMsg.Payload.PingPong == "PING" {
					node.server.SendPingMessage(&serverMsg.Caller, "PONG")
				}
			case Store:
				// serverMsg.Payload.Method == store
				// serverMsg.Payload.Key
				// serverMsg.payload.value
				// node.Datastore.save(key,value)
			case FindNode:
				if serverMsg.Payload.Candidates == nil {
					node.ReturnCandidates(&serverMsg.Caller, &serverMsg.Payload.FindNode)
				} else {
					node.outgoingRequests.mutex.Lock()
					if node.outgoingRequests.outgoingRegister[*serverMsg.Caller.ID] > 0 {
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
				// key = msg.Payload.Key
				// res = node.DataStoreTable.get(key)
				// return res
			default:
				log.Println("PANIC - UNKNOWN RPC METHOD")
			}

		}
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
