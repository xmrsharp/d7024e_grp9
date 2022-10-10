package d7024e

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
	// address := "127.0.0.1:8888" //tror det ska vara 127.0.0.(0 eller 1)
	// con, err := net.Dial("udp", address)
	//
	fmt.Println("starting CLI")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		parseInput := strings.TrimSpace(input)

		if parseInput == "" {
			continue
		} else {
			commands := strings.Fields(parseInput)
			Commands(output, node, commands)
		}
	}
}
