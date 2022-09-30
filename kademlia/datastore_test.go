package d7024e

import (
	"D7024E_GRP9/kademlia/datastore"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInsert(t *testing.T) {
	var data datastore.DataStore

	//Test getting a specific key
	data = datastore.New()
	val := "testar"
	log.Println("Inserting value (calling insert)")
	data.Insert(val)
	log.Println("Testing getting the inserted value...")
	assert.Equal(t, data.Get(NewKademliaID(&val)), "testar")

	//Test getting non existing key, should fail.
	log.Println("Test getting a value that doesn't exist...")
	data = datastore.New()
	val = "testar2"
	assert.Equal(t, data.Get(NewKademliaID(&val)), "")

}
