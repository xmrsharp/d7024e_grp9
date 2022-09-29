package datastore

import (
	"D7024E_GRP9/kademlia/kademliaid"
	
)

type KademliaMap = map[kademliaid.KademliaID]string

type DataStore struct {
	data KademliaMap
}

func New() DataStore {
	return DataStore{make(KademliaMap)}
}

func (data *DataStore) Insert(val string) {
	id := kademliaid.NewKademliaID(&val)
	data.data[id] = val
}

func (data *DataStore) Get(key kademliaid.KademliaID) string {
	return data.data[key]
}

var Store = New()