package d7024e

type KademliaMap = map[KademliaID][20]byte

type DataStore struct {
	data KademliaMap
}

func NewDataStore() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(key KademliaID, val [20]byte) {
	//id := NewKademliaID(val)
	data.data[key] = val
}

func (data *DataStore) Get(key KademliaID) [20]byte {
	return data.data[key]
}
