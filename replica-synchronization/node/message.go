package node

// implement acknowledgement message and timeout

type Message struct {
	Type          string // DISCOVER | SYNC | ACK
	NodeId        int
	Payload       []int        // replica
	ClientList    map[int]string // Ring structure
	Ring          []int
	CoordinatorId int
}

func (m Message) IsEmpty() bool {
	return m.Type == "" && m.NodeId == 0 && m.Payload == nil && m.ClientList == nil && m.Ring == nil
}
