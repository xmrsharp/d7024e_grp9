package kademlia

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var in *os.File = os.Stdin
var out io.Writer = os.Stdout

func Cli(output io.Writer, node *Kademlia) {
	fmt.Print(">> ")
	reader := bufio.NewReader(in)
	for {
		input, _ := reader.ReadString('\n')
		parseInput := strings.TrimSpace(input)

		if parseInput == "" {
			go Cli(output, node)
			return
		} else {
			commands := strings.Fields(parseInput)
			Commands(output, node, commands)
		}
	}
}
