package kademlia

import (
	"fmt"
	"log"
)

type KademliaMap = map[KademliaID]string

type DataStore struct {
	data KademliaMap
}

func NewDataStore() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(key KademliaID, val string) {
	data.data[key] = val
}

func (data *DataStore) KeyExist(key KademliaID) bool {
	_, ok := data.data[key]
	return ok
}

func (data *DataStore) Get(key KademliaID) string {
	if val, ok := data.data[key]; ok {
		log.Printf("GET in DATASTORE found VAL: %s", string(val))
		return val
	} else {
		return ""
	}
}

func (data *DataStore) PrintStore(test KademliaMap) {
	fmt.Println("CURRENT DATASTORE:")
	for key, value := range data.data {
		fmt.Printf("[%s]:[%s]", key, value)
	}
}
