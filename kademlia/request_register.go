package kademlia

import (
	"log"
	"sync"
)

// OutgoingRegister used by Kademlia and network to synchronize expected incoming responses by other nodes.
type OutgoingRegister struct {
	mutex    *sync.Mutex
	register map[KademliaID]int
}

func NewOutgoingRegister() *OutgoingRegister {
	return &OutgoingRegister{&sync.Mutex{}, make(map[KademliaID]int)}
}

func (reg *OutgoingRegister) RegisterOutgoing(target KademliaID) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.register[target] += 1
}

func (reg *OutgoingRegister) RegisterIncoming(target KademliaID) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	if reg.register[target] < 1 {
		log.Println("PANIC - PROCESSED REQUEST FROM UNKNOWN.")
	}
	reg.register[target] -= 1

}

func (reg *OutgoingRegister) ExpectingRequest(target KademliaID) bool {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	if reg.register[target] > 0 {
		reg.register[target] -= 1
		return true
	}
	return false

}

func (reg *OutgoingRegister) ExpectingAnyRequest() bool {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for _, v := range reg.register {
		if v > 0 {
			return true
		}
	}
	return false

}
