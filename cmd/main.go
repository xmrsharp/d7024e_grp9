package main

import (
	"D7024E_GRP9/network"
	"fmt"
	"log"
)

func main() {
	// inc channel from network.
	inc_channel := make(chan string)
	// outgoign channel to network.
	out_channel := make(chan string)
	server_test := network.InitNetwork("0.0.0.0", 8888, inc_channel, out_channel)
	// Try at fixing channels
	// 	kademlia_node := Kademlia{}
	// 	kademlia_node.run()
	fmt.Println("Calling go routine.")
	go server_test.Listen()
	for {
		fmt.Println("Calling channel")
		fmt.Println("no need to wait.")
		msg_from_server := <-inc_channel
		log.Println("RECIEVED FROM CHANNEL:", msg_from_server)
	}

}
