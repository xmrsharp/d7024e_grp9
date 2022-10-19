package kademlia

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetInsert(t *testing.T) {
	//Test getting a specific key
	data := NewDataStore()
	val := "testar"
	key := NewKademliaID(&val)
	log.Println("Inserting value (calling insert)")
	data.Insert(key, val)
	log.Println("Testing getting the inserted value...")
	assert.Equal(t, data.Get(key), "testar")

	//Test getting non existing key, should fail.
	log.Println("Test getting a value that doesn't exist...")
	data = NewDataStore()
	val = "testar2"
	assert.Equal(t, data.Get(NewKademliaID(&val)), "")

}
