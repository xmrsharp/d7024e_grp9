package kademlia

import (
	"D7024E_GRP9/api"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	ALPHA_VALUE = 3
	K_VALUE     = 20
)

// TODO go over all channels in play.
// Logic and state of Kademlia node
type Kademlia struct {
	kademliaServer      *Network
	apiServer           *api.APIServer
	outgoingRequests    *OutgoingRegister
	channelAPI          <-chan api.APIChannel
	channelServerInput  <-chan msg
	channelServerOutput chan<- msg
	channelNodeLookup   chan []Contact
	routingTable        *RoutingTable
	datastore           DataStore
	channelDataStore    chan []byte
}

func NewKademlia(ip string, portKademlia int, portAPI int, id *KademliaID) *Kademlia {

	// Init shared channels
	channelServerInput := make(chan msg)
	channelServerOutput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	channelDataStore := make(chan []byte)

	addrs := ip + ":" + strconv.Itoa(portKademlia)
	outgoingRequests := NewOutgoingRegister()
	selfContact := NewContact(id, addrs)
	kademliaServer := NewNetwork(selfContact, addrs, outgoingRequests, channelServerInput, channelServerOutput)
	routingTable := NewRoutingTable(selfContact)

	channelAPI := make(chan api.APIChannel)
	apiServer := api.NewServer(ip, portAPI, channelAPI)

	datastore := NewDataStore()

	kademliaNode := Kademlia{kademliaServer, apiServer, outgoingRequests, channelAPI, channelServerInput, channelServerOutput, channelNodeLookup, routingTable, datastore, channelDataStore}
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
			if consumedCandidates[tempContact.ID] == 1 || *tempContact.ID == *node.routingTable.me.ID {
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
		log.Println("At newcandidates channel row 107")
		newCandidates = <-node.channelNodeLookup
		// Investigate to why I have to recalculate the distances?
		log.Println("After newcandidates")
		for i := 0; i < len(newCandidates); i++ {
			newCandidates[i].CalcDistance(target)
			node.kademliaServer.SendPingMessage(&(newCandidates[i]))
		}
		log.Println("After forloop row 111")
		// Only need to check head of list as list is ordered on arrival.
		if !currentClosestNode.Less(&newCandidates[0]) {
			// Got closer to target, update current closest and succesfull probes.
			probedNoCloser = 0
			currentClosestNode = newCandidates[0]
			log.Println("inside if on 117")
		} else {
			// No closer to target.
			probedNoCloser++
			log.Println("Inside else on 125")
		}
		log.Println("Append currcand and subtract aAC")
		currentCandidates.Append(newCandidates)
		activeAlphaCalls--
	}
}

func (node *Kademlia) FindClosestContacts(target *KademliaID, count int) {
	node.NodeLookup(target)
}

/*
	 	Check that the hash is valid
	 	Check if the data is found on this node
		Send request to look for the data on other nodes
			NodeLookup (key = target)
			Send request to all contacts found in the lookup.
*/
func (node *Kademlia) LookupData(key KademliaID) []byte {
	log.Println("LookupData command in kademlia.go called with key:")
	log.Println(&key)
	var val []byte
	if node.datastore.Get(key) != nil {
		val = <-node.channelDataStore
		return val
	} else {
		node.NodeLookup(&key)

		//Array with contacs
		neighbours := node.routingTable.FindClosestContacts(&key, BucketSize)
		log.Println("After neighbours in lookupdata")
		// skicka till alla noder  fan yolo
		// mao node lookup sendfinddata till alla du hittat
		for i := 0; i < len(neighbours); i++ {
			go node.kademliaServer.SendFindDataMessage(&(neighbours[i]), key)
			log.Println("Trying to find data on node: " + (neighbours[i]).ID.String())
		}
		return nil
	}

}

/*
Create new random ID
Update routing table for the new id
Get closest contacts for id and try sending store message to them
*/

func (node *Kademlia) StoreValue(data []byte) {
	log.Println("Store command in kademlia.go called with data:")
	log.Println(data)
	hashed := Hash(data)
	//str2B := []byte(hashed)
	key := NewKademliaID(&hashed)

	//contacts := kademlia.LookupContact((node.KademliaID)(str2B))
	log.Println("Trying to store key: " + key.String())
	node.NodeLookup(&key)
	log.Println("Nodelookup complete, neighbour next")
	//Array with contacs
	neighbours := node.routingTable.FindClosestContacts(&key, K_VALUE)
	//Loop through closest contacts and try storing on them.
	for i := 0; i < len(neighbours); i++ {
		storeStatus := make(chan bool)
		go node.kademliaServer.SendStoreMessage(&(neighbours[i]), key, data, storeStatus)
		select {
		case <-storeStatus:
			log.Println("Stored data on node: " + (neighbours[i]).ID.String())
		case <-time.After(10 * time.Second):
			log.Println("TOOK TOO LONG TIME!!")
		}
	}
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
	node.NodeLookup(node.routingTable.me.ID)

}

func (node *Kademlia) handleIncomingRPC(kademliaServerMsg msg) {
	log.Printf(("RECIEVED [%s] EVENT FROM [%s]:[%s]"), kademliaServerMsg.Method.String(), kademliaServerMsg.Caller.Address, kademliaServerMsg.Caller.ID)
	if kademliaServerMsg.Caller.Address == node.routingTable.me.Address {
		log.Printf("SUSPECT CALLER [%s] REASON: IDENTICAL ADDRS AS SERVER", kademliaServerMsg.Caller.ID)
		return
	}

	// Adding caller of any incoming request.
	node.routingTable.AddContact(kademliaServerMsg.Caller)
	switch kademliaServerMsg.Method {
	case Ping:
		if kademliaServerMsg.Payload.PingPong == "PING" {
			node.kademliaServer.SendPongMessage(&kademliaServerMsg.Caller)
		}
	case FindNode:
		// TODO Need to check that register is correctly incremented/decremented.
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
		// key = msg.Payload.Key
		// res = node.DataStoreTable.get(key)
		// return res
		if kademliaServerMsg.Payload.Value == nil {
			//ngn vill hitta värdet
			val := node.datastore.Get(kademliaServerMsg.Payload.Key)
			node.kademliaServer.SendReturnDataMessage(&kademliaServerMsg.Caller, val)
		} else {
			// svar från annan nod datastore att här är värdet
			node.channelDataStore <- kademliaServerMsg.Payload.Value
		}
	case Store:
		// TODO Handle inc store event.
		key := kademliaServerMsg.Payload.Store.Key
		value := kademliaServerMsg.Payload.Store.Value
		node.datastore.Insert(key, value)
	default:
		log.Println("PANIC - UNKNOWN RPC METHOD")
	}
}

func (node *Kademlia) handleIncomingAPIRequest(apiRequest api.APIChannel) {
	switch apiRequest.ApiRequestMethod {
	case "GET_VALUE":
		key := apiRequest.ApiRequestPayload
		log.Println("INSERT CALL HERE TO GET VALUE FROM KEY:", key, " AND RETURN THE VALUE")
		apiRequest.ApiResponseChannel <- []byte("RESPONSE FROM GET VALUE")
	case "STORE_VALUE":
		valueStore := apiRequest.ApiRequestPayload
		log.Println("INSERT CALL HERE TO STORE VALUE OF:", valueStore, " AND RETURN THE KEY HASH FROM CREATED KEY VALUE")
		apiRequest.ApiResponseChannel <- []byte("RESPONSE FROM STORE VALUE")
	default:
		log.Panic("RECIEVED INVALID API REQUEST METHOD FROM HTTP SERVER", apiRequest)

	}

}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	go node.kademliaServer.Listen()
	go node.apiServer.Listen()
	go node.bootLoader(bootLoaderAddrs, bootLoaderID)
	go Cli(os.Stdout, node)
	for {
		select {
		case apiRequest := <-node.channelAPI:
			node.handleIncomingAPIRequest(apiRequest)
		case kademliaServerMsg := <-node.channelServerInput:
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
