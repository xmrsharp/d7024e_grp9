package kademlia

import "log"

type KademliaMap = map[KademliaID]string

type DataStore struct {
	data KademliaMap
}

func NewDataStore() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(key KademliaID, val string) {
	log.Printf("INSERT in DATASTORE called with KEY %s, VAL %s: ", key.String(), val)
	//id := NewKademliaID(val)
	data.data[key] = val
	// data.PrintStore(data.data)
}

func (data *DataStore) KeyExist(key KademliaID) bool {
	_, ok := data.data[key]
	return ok
}

func (data *DataStore) Get(key KademliaID) string {
	log.Printf("GET in DATASTORE called with KEY: %s", key.String())
	if val, ok := data.data[key]; ok {
		log.Printf("GET in DATASTORE found VAL: %s", string(val))
		return val
	} else {
		return ""
	}
}

func (data *DataStore) PrintStore(test KademliaMap) {
	// loop over elements of slice
	log.Printf("CURRENT DATASTORE:")
	for key, value := range data.data {
		log.Printf("[%s]:[%s]", key, value)
	}
}
