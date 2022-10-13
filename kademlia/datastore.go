package kademlia

type KademliaMap = map[KademliaID][]byte

type DataStore struct {
	data KademliaMap
}

func NewDataStore() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(key KademliaID, val []byte) {
	//id := NewKademliaID(val)
	data.data[key] = val
}

func (data *DataStore) Get(key KademliaID) []byte {
	if val, ok := data.data[key]; ok {
		return val
	} else {
		return nil
	}
}
