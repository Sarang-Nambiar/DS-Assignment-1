package node

import (
	// "fmt"
)

// implement acknowledgement message and timeout

type Message struct{
	Type string // PUT | SYNC
	nodeId int
	Payload string // replica
}

func (m Message) IsEmpty() bool {
	return m.Type == "" && m.nodeId == 0 && m.Payload == ""
}

