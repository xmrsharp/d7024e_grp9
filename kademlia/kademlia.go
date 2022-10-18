package kademlia

import (
	"D7024E_GRP9/api"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// TODO Address constant value missuse
const (
	ALPHA_VALUE = 3
	K_VALUE     = 20
)

// TODO Discuss about changing go version from 1.13 -> 1.18
// Logic and state of Kademlia node
type Kademlia struct {
	kademliaServer     *Network
	apiServer          *api.APIServer
	outgoingRequests   *OutgoingRegister
	channelAPI         <-chan api.APIChannel // api channel.
	channelServerInput <-chan msg            // kad node server channel.
	channelNodeLookup  chan []Contact        // NodeLookup rpc
	channelStoreValue  chan KademliaID       // Store channel rpc
	channelValueLookup chan msg              // FindValue channel rpc
	datastore          DataStore
	routingTable       *RoutingTable
}

func NewKademlia(ip string, portKademlia int, portAPI int, id *KademliaID) *Kademlia {

	// Init shared channels
	channelServerInput := make(chan msg)
	channelNodeLookup := make(chan []Contact)
	channelStoreValue := make(chan KademliaID)
	channelValueLookup := make(chan msg)

	addrs := ip + ":" + strconv.Itoa(portKademlia)
	outgoingRequests := NewOutgoingRegister()
	selfContact := NewContact(id, addrs)
	kademliaServer := NewNetwork(selfContact, addrs, outgoingRequests, channelServerInput)
	routingTable := NewRoutingTable(selfContact)

	channelAPI := make(chan api.APIChannel)
	apiServer := api.NewServer(ip, portAPI, channelAPI)

	datastore := NewDataStore()

	kademliaNode := Kademlia{kademliaServer: kademliaServer,
		apiServer:          apiServer,
		outgoingRequests:   outgoingRequests,
		channelAPI:         channelAPI,
		channelServerInput: channelServerInput,
		channelNodeLookup:  channelNodeLookup,
		channelStoreValue:  channelStoreValue,
		channelValueLookup: channelValueLookup,
		routingTable:       routingTable,
		datastore:          datastore}
	return &kademliaNode
}

func (node *Kademlia) ReturnCandidates(caller *Contact, target *KademliaID) {
	candidates := node.routingTable.FindClosestContacts(target, 20)
	node.kademliaServer.respondFindContactMessage(caller, candidates)
}

func (node *Kademlia) NodeLookup(target *KademliaID) {
	func(drain chan []Contact) {
		for {
			select {
			case _ = <-drain:
				log.Println("REMOVED STALE FIND_NODE RESPONSE")
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
		case <-time.After(3 * time.Second):
			log.Println("Failed to recieve FIND_NODE response within time interval")
			return
		}
	}
}

func (node *Kademlia) LookupData(key KademliaID) Result {
	val := node.datastore.Get(key)
	var response Result
	if val != "" {
		response.ID = *node.routingTable.me.ID
		response.Value = val
		return response // Works, meaning found value in local node.
	}

	// Refresh any previously stored values in pipe.
	func() {
		for {
			select {
			case _ = <-node.channelValueLookup:
				log.Println("STALE GET VALUE FLUSHED")
			default:
				return
			}
		}
	}()
	node.NodeLookup(&key)
	neighbours := node.routingTable.FindClosestContacts(&key, BucketSize)
	for i := 0; i < len(neighbours); i++ {
		if *neighbours[i].ID != *node.routingTable.me.ID {
			node.kademliaServer.SendFindDataMessage(&(neighbours[i]), key)
		}
	}

	var responseAck Result
	for {
		select {
		case valueConfirmation := <-node.channelValueLookup:
			proposedValue := valueConfirmation.Payload.FindValue.Value
			if NewKademliaID(&proposedValue) == key {
				responseAck.ID = *valueConfirmation.Caller.ID
				responseAck.Value = proposedValue
				return responseAck
			}
		case <-time.After(3 * time.Second):
			responseAck.Err = errors.New("Failed to fetch value of key")
			return responseAck
		}
	}
}

func (node *Kademlia) StoreValue(data string) Result {
	key := NewKademliaID(&data)

	// Unloading all old accepted store values from channel.
	func() {
		for {
			select {
			case _ = <-node.channelStoreValue:
				log.Println("STALE STORE ACK FLUSHED")
			default:
				return
			}
		}
	}()

	node.NodeLookup(&key)
	storeCandidates := node.routingTable.FindClosestContacts(&key, K_VALUE)
	// NOTE 2 For testing purposes
	//for i := 0; i < len(storeCandidates); i++ {
	for i := 0; i < 2; i++ {
		if *storeCandidates[i].ID != *node.routingTable.me.ID {
			node.kademliaServer.SendStoreMessage(&(storeCandidates[i]), key, data)
		}
	}

	var responseAck Result
	for {
		select {
		case keyConfirmation := <-node.channelStoreValue:
			if keyConfirmation == key {
				responseAck.ID = key
				return responseAck
			}
		case <-time.After(3 * time.Second):
			responseAck.Err = errors.New("Failed to ack store value.")
			return responseAck
		}
	}

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
			node.kademliaServer.respondPingMessage(&kademliaServerMsg.Caller)
		}
	case FindNode:
		go func() {
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
		}()
	case FindValue:
		go func() {
			if kademliaServerMsg.Payload.FindValue.Value == "" {
				// Call to search for value in own datastore
				val := node.datastore.Get(kademliaServerMsg.Payload.FindValue.Key)
				if val != "" {
					node.kademliaServer.respondFindDataMessage(&kademliaServerMsg.Caller, val)
				}
			} else {
				// Response from other node of value.
				node.channelValueLookup <- kademliaServerMsg
			}
		}()
	case Store:
		go func() {
			if kademliaServerMsg.Payload.Store.Value == "" {
				node.channelStoreValue <- kademliaServerMsg.Payload.Store.Key
				return
			}
			key := kademliaServerMsg.Payload.Store.Key
			value := kademliaServerMsg.Payload.Store.Value
			node.datastore.Insert(key, value)
			if node.datastore.KeyExist(key) {
				node.kademliaServer.respondStoreMessage(&kademliaServerMsg.Caller, key)
			}
		}()

	default:
		log.Println("PANIC - UNKNOWN RPC METHOD")
	}
}

func (node *Kademlia) handleIncomingAPIRequest(apiRequest api.APIChannel) {
	var res *Result
	switch apiRequest.ApiRequestMethod {
	case "GET_VALUE":
		key := apiRequest.ApiRequestPayload
		var id KademliaID
		for i := 0; i < 20; i++ {
			id[i] = key[i]
		}
		log.Println("LOOKING UP KEY: ", id)
		*res = node.LookupData(id)
		log.Println("GET_VALUE: RECIEVED:", *res)
		log.Println("INSERT CALL HERE TO GET VALUE FROM KEY:", key, " AND RETURN THE VALUE")
	case "STORE_VALUE":
		valueToStore := apiRequest.ApiRequestPayload
		*res = node.StoreValue(string(valueToStore))
		log.Println("STORE_VALUE RECIEVED::", *res)
	default:
		log.Panic("RECIEVED INVALID API REQUEST METHOD FROM HTTP SERVER", apiRequest)

	}
	go func(res Result, method string, resp chan []byte) {
		log.Println("AND RES IN GO ROUTINE IS :", res)
		if res.IsError() {
			log.Println("ERROR:", method)
			log.Println(res)
			resp <- []byte("ERROR")
			return
		}
		resp <- []byte(res.ID.String() + " - " + res.Value)
	}(*res, apiRequest.ApiRequestMethod, apiRequest.ApiResponseChannel)
}

func (node *Kademlia) Run(bootLoaderAddrs string, bootLoaderID KademliaID) {
	// ERROR IF BOOT LOADER SOMEHOW STARTS BEFORE NETWORK.
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
