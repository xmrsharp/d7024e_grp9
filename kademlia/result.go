package kademlia

type Result struct {
	ID    KademliaID // Treated as node id and key digest.
	Value string
	Err   error
}

func (r *Result) IsError() bool {
	return r.Err != nil
}
