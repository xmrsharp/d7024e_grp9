package kademlia

import (
	"fmt"
	"io"
	"log"
	"os"
)

func Commands(output io.Writer, node *Kademlia, commands []string) {
	switch commands[0] {
	case "put":
		if len(commands) == 2 {
			Put(node, commands[1])
		} else {
			fmt.Println("Arg error")
		}
	case "get":
		if len(commands) == 2 {
			Get(node, commands[1])
		} else {
			fmt.Println("Arg error")
		}
	case "exit":
		os.Exit(0)
	default:
		fmt.Println(output, "invalid command")
	}
}

// TODO PUT is currently being treated as POST - not idempotent.
func Put(node *Kademlia, input string) {
	//str2B := []byte(input)

	res := node.StoreValue(input)
	log.Println("RECIEVED RES : ", res)
	//fmt.Println("Hash is: ", str2B)
}
func Get(node *Kademlia, hash string) {
	//str2B := []byte(hash)
	key := NewKademliaID(&hash)
	value := node.LookupData(key)
	fmt.Printf("GET CALLED WITH KEY: %s", key.String())
	stringValue := string(value[:])
	fmt.Println("STRINGVALUE IS: " + stringValue)
	if stringValue == "" {
		fmt.Println("Couldn't find requested value")
	} else {
		fmt.Println("Value found: ", stringValue, " In node: ", key.String())
	}
}
