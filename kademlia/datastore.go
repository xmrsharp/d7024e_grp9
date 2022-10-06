package kademlia

type KademliaMap = map[KademliaID]string

type DataStore struct {
	data KademliaMap
}

func NewDataStore() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(val *string) {
	id := NewKademliaID(val)
	data.data[id] = *val
}

func (data *DataStore) Get(key KademliaID) string {
	return data.data[key]
}
