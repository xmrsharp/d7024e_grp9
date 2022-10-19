package tests_tests

import (
	"D7024E_GRP9/kademlia"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

//	"D7024E_GRP9/kademlia/datastore"
func TestGetInsert(t *testing.T) {
	//Test getting a specific key
	data := kademlia.NewDataStore()
	val := "testar"
	key := kademlia.NewKademliaID(&val)
	log.Println("Inserting value (calling insert)")
	data.Insert(key, val)
	log.Println("Testing getting the inserted value...")
	assert.Equal(t, data.Get(key), "testar")

	//Test getting non existing key, should fail.
	log.Println("Test getting a value that doesn't exist...")
	data = kademlia.NewDataStore()
	val = "testar2"
	assert.Equal(t, data.Get(kademlia.NewKademliaID(&val)), "")

}
