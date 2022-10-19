package kademlia

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// kallas för att testa kommandon, commands.go kod ska skickas till output efter Commands() kallas, försöker läsa av det som skrivs av command och returnar det
// kanske behöver skapa en test node och byta ut nil i Commands med test noden
func comSend(com string) string {
	output := bytes.NewBuffer(nil)
	comArray := strings.Fields(com)
	Commands(output, nil, comArray)
	return removeNewline(output)
}

func removeNewline(out io.Writer) string {
	str := out.(*bytes.Buffer).String()
	return strings.TrimSuffix(str, "\n")
}

func TestCommandGet(t *testing.T) {
	assert.Equal(t, "", comSend("get")) //skicka fel antal inputs, som ska returna "Arg error"
}

func TestCommandPut(t *testing.T) {
	assert.Equal(t, "", comSend("put")) //skicka fel antal inputs, som ska returna "Arg error"
}

/*
	 func TestCommandExit(t *testing.T) {
		Behöver hitta sätt att testa exot code
	}
*/
func TestWrongCommand(t *testing.T) {
	assert.Equal(t, "", comSend("test"))
}

/* func TestGetWrong(t *testing.T) {
	assert.Equal(t, "Did not find value test", comSend("get test"))
}
*/
//Går inte att köra PGA nil node i comSend, kanske pröva med fake node
