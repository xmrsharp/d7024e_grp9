package d7024e

import (
	"fmt"
	"io"
	"os"
)

func Commands(output io.Writer, node *Kademlia, commands []string) {
	switch commands[0] {
	case "put":
		if len(commands) == 2 {
			Put(*node, commands[1])
		} else {
			fmt.Println("Arg error")
		}
	case "get":
		if len(commands) == 2 {
			Get(*node, commands[1])
		} else {
			fmt.Println("Arg error")
		}
	case "exit":
		os.Exit(0)
	default:
		fmt.Println(output, "invalid command")
	}
}

func Put(node Kademlia, input string) {
	hash := node.Store(&input)
	fmt.Println("Hash is: ", hash)
}
func Get(node Kademlia, hash string) {
	value, id := node.LookupData(hash)
	if id == "" {
		fmt.Println("Couldn't find requested value")
	} else {
		fmt.Println("Value found: ", value, " in node: ", id)
	}
}
