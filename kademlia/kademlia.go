package kademlia

import (
	"D7024E_GRP9/api"
	"fmt"
	"log"
	"strconv"
)

const (
	ALPHA_VALUE = 3
)

// Logic and state of Kademlia node
type Kademlia struct {
	kademliaServer      *Network
	apiServer           *api.APIServer
	outgoingRequests    *OutgoingRegister
	channelAPI          chan string
	channelServerInput  <-chan msg
	channelServerOutput chan<- msg
	channelNodeLookup   chan []Contact
	routingTable        *RoutingTable
	datastore           DataStore
}

func NewKademlia(ip string, portKademlia int, portAPI int, id *KademliaID) *Kademlia {
	channelServerInput := make(chan msg)
	channelServerOutput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	channelAPI := make(chan string)
	addrs := ip + ":" + strconv.Itoa(portKademlia)
	outgoingRequests := NewOutgoingRegister()
	selfContact := NewContact(id, addrs)
	kademliaServer := NewNetwork(selfContact, addrs, outgoingRequests, channelServerInput, channelServerOutput)
	apiServer := api.NewServer(ip, portAPI, channelAPI)
	routingTable := NewRoutingTable(selfContact)
	datastore := NewDataStore()

	kademliaNode := Kademlia{kademliaServer, apiServer, outgoingRequests, channelAPI, channelServerInput, channelServerOutput, channelNodeLookup, routingTable, datastore}
	return &kademliaNode
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.kademliaServer.respondFindContactMessage(caller, candidates)
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
				node.kademliaServer.SendFindContactMessage(&tempContact, *target)
			}
			activeAlphaCalls++
		}
		if !node.outgoingRequests.ExpectingAnyRequest() {
			return
		}
		newCandidates = <-node.channelNodeLookup
		// Investigate to why I have to recalculate the distances?
		for i := 0; i < len(newCandidates); i++ {
			newCandidates[i].CalcDistance(target)
			node.kademliaServer.SendPingMessage(&(newCandidates[i]))
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

func (node *Kademlia) LookupData(hash string) {
	// TODO
}

func (node *Kademlia) Store(data *string) {
	node.datastore.Insert(data)
}

func (node *Kademlia) bootLoader(bootLoaderAddrs string, bootLoaderID KademliaID) {
	if bootLoaderAddrs == "" {
		return
	}
	bootContact := NewContact(&bootLoaderID, bootLoaderAddrs)
	node.routingTable.AddContact(bootContact)
	node.NodeLookup(node.routingTable.me.ID)
}

func (node *Kademlia) handleIncomingRPC(kademliaServerMsg msg) {
	log.Printf(("RECIEVED [%s] EVENT FROM [%s]:[%s]"), kademliaServerMsg.Method.String(), kademliaServerMsg.Caller.Address, kademliaServerMsg.Caller.ID)
	if kademliaServerMsg.Caller.Address == node.routingTable.me.Address {
		log.Printf("SUSPECT CALLER [%s] REASON: IDENTICAL ADDRS AS SERVER", kademliaServerMsg.Caller.ID)
		return
	}
	node.routingTable.AddContact(kademliaServerMsg.Caller)
	switch kademliaServerMsg.Method {
	case Ping:
		if kademliaServerMsg.Payload.PingPong == "PING" {
			node.kademliaServer.SendPongMessage(&kademliaServerMsg.Caller)
		}
	case FindNode:
		// TODO EVALUATE IF REGISTER IS CORRECTLY IMPLEMENTED WITH COUNTERS.
		if kademliaServerMsg.Payload.Candidates == nil {
			node.ReturnCandidates(&kademliaServerMsg.Caller, &kademliaServerMsg.Payload.FindNode)
		} else {
			if node.outgoingRequests.ExpectingRequest(*kademliaServerMsg.Caller.ID) {
				node.channelNodeLookup <- kademliaServerMsg.Payload.Candidates
			} else {
				log.Printf("SUSPECT CALLER [%s] REASON: UNEXPECTED FIND_NODE RESPONSE", kademliaServerMsg.Caller.ID)
			}
		}
	case FindValue:
		// TODO Handle inc find value event.
	case Store:
		// TODO Handle inc store event.
	default:
		log.Println("PANIC - UNKNOWN RPC METHOD")
	}
}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	go node.kademliaServer.Listen()
	go node.apiServer.Listen()
	go node.bootLoader(bootLoaderAddrs, bootLoaderID)
	for {
		log.Println("STARTING AGAIN")
		select {
		case apiMsg := <-node.channelAPI:
			log.Println("RECIEVED API REQUEST: ", apiMsg)
			node.channelAPI <- "RESPONSE API"
		case kademliaServerMsg := <-node.channelServerInput:
			log.Println("RECIEVED KAD NODE REQUEST")
			node.handleIncomingRPC(kademliaServerMsg)
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
