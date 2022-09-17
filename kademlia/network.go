package d7024e

// Plan:
//  - unique identifier.
//  - should every ip simply be 0.0.0.0??
//  - If routing table is empty -> pass arg to connect to.
import (
    "fmt"
    "net"
    "sync"
)

// Strategy:
//  - Should ALWAYS be listening for incoming connections.
//  - When connecting to other servers -> make request and go routine on response.
type Network struct {
    wg sync.WaitGroup
    self_id KademliaID // storing pointer to id here as we need to send our id to any caller etc.
}

func (network *Network) Listen(ip string, port int) {
    server_socket, err := net.Listen("udp",ip+":"+port)
    if err != nil{
        fmt.Println("failed to bind server on: ",ip+":"+port,"\nReason: ",err);
    }
    // Order matters? 
    defer network.wg.Wait()
    defer server_socket.Close()
    for{
        conn, err := server_socket.Accept()
        if err != nil{
            fmt.Println("error dealing with conn: ",err);
            break;
        }else{
            // Conn accepted.
            network.wg.Add(1)
            go handleRequest(conn)
        }
    }
}

func handleRequest(conn net.Conn){
    // Note, udp packets are
    headers := make([]byte, 8)
    n_bytes, err := conn.Read(headers)
    fmt.Println("Recieved following headers:",string(headers));
    buff := make([]byte, 8)
    n_bytes, err := conn.Read(buff) 
}

// Inc contant -> (ID, IP ADDRESS, DISTANCE.)
func (network *Network) SendPingMessage(contact *Contact) {
	// Simply ping ip address and append our contact info.
    //  
}

// These methods does not return anything? dacuc?
// Return contact or list of contacts?
func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}


// Return wrapper of contact lists or data.
func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

// Return nothing as we're simply passing data to others to handle
// Or for us to store if we're closer/closest.
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
