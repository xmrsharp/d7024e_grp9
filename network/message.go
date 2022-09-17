package network


// Have to keep attributes exported as encoding/gob package requires exported struct fields
type msg struct {
	Method  rpc_method
	Payload string
}

// Careful with this 'enumeration' method, simple integers will work aswell.
type rpc_method int

const (
	Ping rpc_method = iota
	Store
	FindData
	FindNode
)
