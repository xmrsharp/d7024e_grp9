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
	channelDataStore    chan string
}

func NewKademlia(ip string, portKademlia int, portAPI int, id *KademliaID) *Kademlia {

	// Init shared channels
	channelServerInput := make(chan msg)
	channelServerOutput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	channelDataStore := make(chan string)

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
	// Draining stale responses from channel (NOTE WIP)
	func(drain chan []Contact) {
		for {
			select {
			case _ = <-drain:
				log.Println("REMOVED STALE FIND_NODE RESPONSE:")
			default:
				return
			}
		}
	}(node.channelNodeLookup)

	// Clean out any straglers from incoming requests that have not responded within a set amount of time.
	defer node.outgoingRequests.ResetRegister()
	// Initiate candidates.
	currentCandidates := ContactCandidates{node.routingTable.FindClosestContacts(target, BucketSize)}

	// Inititate end condition.
	currentClosestNode := currentCandidates.GetClosest()
	probedNoCloser := 0

	// Visited nodes
	consumedCandidates := make(map[KademliaID]int)

	// For keeping track off active outgoing FindNode RPCs.
	activeAlphaCalls := 0
	for probedNoCloser < BucketSize {
		if currentClosestNode.ID.Equals(target) {
			return
		}
		for activeAlphaCalls < ALPHA_VALUE && currentCandidates.Len() > 0 {
			tempContact := currentCandidates.PopClosest()
			// Check for alrdy consumed nodes.
			if consumedCandidates[*tempContact.ID] == 1 || *tempContact.ID == *node.routingTable.me.ID {
				// Already consumed contact - dead end.
				activeAlphaCalls--
			} else {
				consumedCandidates[*tempContact.ID] += 1
				node.kademliaServer.SendFindContactMessage(&tempContact, *target)
			}
			activeAlphaCalls++
		}
		if !node.outgoingRequests.ExpectingAnyRequest() {
			return
		}
		select {
		case newCandidates := <-node.channelNodeLookup:
			// Recieved response in allowed interval
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
		case <-time.After(time.Second):
			log.Println("Failed to recieve FIND_NODE response within time interval")
			return
		}
	}
}

/*
	 	Check that the hash is valid
	 	Check if the data is found on this node
		Send request to look for the data on other nodes
			NodeLookup (key = target)
			Send request to all contacts found in the lookup.
*/
func (node *Kademlia) LookupData(key KademliaID) string {
	log.Println("LookupData command in kademlia.go called with key:")
	log.Println(&key)
	val := node.datastore.Get(key)
	log.Printf("VAL is %s AFTER datastore GET with KEY %s", val, key.String())

	if val != "" {
		return val
	}
	node.NodeLookup(&key)
	neighbours := node.routingTable.FindClosestContacts(&key, BucketSize)

	log.Println("After neighbours in lookupdata")
	for i := 0; i < len(neighbours); i++ {
		log.Printf("Forloop in LOOKUPDATA on lap: %d", i)
		if !(neighbours[i].ID.Equals(node.routingTable.me.ID)) {
			node.kademliaServer.SendFindDataMessage(&(neighbours[i]), key)
		} else {
			log.Println("Got call to not run a SENDFINDDATAMESSAGE IN DAVID")
		}
	}

	val = <-node.channelDataStore
	return val
}

/*
Create new random ID
Update routing table for the new id
Get closest contacts for id and try sending store message to them
*/

func (node *Kademlia) StoreValue(data string) {
	log.Println("Store command in kademlia.go called with data:")
	log.Println(data)
	//hashed := Hash(data)
	//str2B := []byte(hashed)
	key := NewKademliaID(&data)

	//contacts := kademlia.LookupContact((node.KademliaID)(str2B))
	log.Println("Trying to store key: " + key.String())
	node.NodeLookup(&key)
	neighbours := node.routingTable.FindClosestContacts(&key, K_VALUE)
	for i := 0; i < len(neighbours); i++ {
		if *neighbours[i].ID != *node.routingTable.me.ID {
			node.kademliaServer.SendStoreMessage(&(neighbours[i]), key, data)
		}
	}

}

func Hash(data string) string {
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
	if kademliaServerMsg.Method != 0 {

		log.Printf(("[%s] RECIEVED [%s] EVENT FROM [%s]:[%s]"), node.routingTable.me.Address, kademliaServerMsg.Method.String(), kademliaServerMsg.Caller.Address, kademliaServerMsg.Caller.ID)
		if kademliaServerMsg.Caller.Address == node.routingTable.me.Address {
			log.Printf("SUSPECT CALLER [%s] REASON: IDENTICAL ADDRS AS SERVER", kademliaServerMsg.Caller.ID)
			return
		}
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
				// NOTE investigate docker compose behaviour - will block (other nodes from responding) until self terminates (socket errors?).
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
		if kademliaServerMsg.Payload.Value == "" {
			//ngn vill hitta värdet
			val := node.datastore.Get(kademliaServerMsg.Payload.Key)
			if val != "" {
				node.kademliaServer.SendReturnDataMessage(&kademliaServerMsg.Caller, val)
			}

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
		valueToStore := apiRequest.ApiRequestPayload
		log.Println("CALL TO STORE VALUE WITH VALUE:", valueToStore)
		node.StoreValue(string(valueToStore))
		log.Println("RETURNING FROM STORE VALUE")
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
