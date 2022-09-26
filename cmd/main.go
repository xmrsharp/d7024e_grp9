package main

import d7024e "D7024E_GRP9/kademlia"

//TODO Rename package structure.

func main() {
	test_node := d7024e.InitKademlia("0.0.0.0", 8888)
	test_node.Run()
}
