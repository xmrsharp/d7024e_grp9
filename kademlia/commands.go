package kademlia

import (
	"fmt"
	"io"
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

func Put(node *Kademlia, input string) {
	res := node.StoreValue(input)
	if res.IsError() {
		fmt.Printf(("FAIL STORE VALUE [%s]"), input)
	} else {
		fmt.Printf("VALUE STORED\n\tVALUE:[%s]\n\tKEY:[%s]", input, res.ID.String())
	}
}
func Get(node *Kademlia, value string) {
	key := NewKademliaID(&value)
	res := node.LookupData(key)
	if res.IsError() {
		fmt.Printf(("DID NOT FIND VALUE [%s]."), value)
	} else {
		fmt.Printf(("FOUND VALUE\n\tNODE:[%s]\n\tKEY:[%s]\n\tVALUE:[%s]\n"), res.ID.String(), key.String(), res.Value)
	}

}
