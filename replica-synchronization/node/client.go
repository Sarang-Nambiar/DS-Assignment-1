package node

import (
	"fmt"
	"net/rpc"
	"slices"
	"strconv"
	"time"
)

type ClientNode struct {
	Node        *Node
	LastUpdated time.Time
}

func (cn *ClientNode) InvokeSynchronization(msg *Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	cn.Node.LocalReplica = msg.Payload
	cn.LastUpdated = time.Now()

	fmt.Printf("[NODE-%d] Replica synchronized with the coordinator. Replica: '%v'\n", cn.Node.Id, msg.Payload)
	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

func (cn *ClientNode) DiscoverRingAndPropagate(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	go cn.propagateToSuccessor(msg)

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

// func (cn *ClientNode) UpdateRingAndPropagate(msg Message, reply *Message) error {
// 	cn.Node.Lock.Lock()
// 	defer cn.Node.Lock.Unlock()

// 	cn.Node.ClientList = msg.ClientList
// 	cn.Node.Ring = msg.Ring
// 	cn.LastUpdated = time.Now()

// 	fmt.Printf("[NODE-%d] Ring structure updated. New ring: %v\n", cn.Node.Id, cn.Node.Ring)

// 	// Propagate to successor
// 	successorId := cn.Node.findSuccessor(msg) // Finding the next node
// 	if successorId != -1 && successorId != msg.NodeId { // Stop if we've reached the coordinator
// 		go cn.propagateToSuccessor(msg)
// 	} else {
// 		fmt.Printf("[NODE-%d] Ring update propagation completed.\n", cn.Node.Id)
// 	}

// 	*reply = Message{
// 		Type: ACK,
// 		NodeId: cn.Node.Id,
// 	}
// 	return nil
// }

func (cn *ClientNode) propagateToSuccessor(msg Message) {
	curId := cn.Node.Id
	done := make(chan error, 1)

	// Only to update the ring structure to include the new node.
	if cn.Node.Id != msg.NodeId {
		cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.findIndex(cn.Node.CoordinatorId), msg.NodeId)// Add the new node to the ring structure
	}

	msg.Ring = append(msg.Ring, cn.Node.Id)
	msg.ClientList[cn.Node.Id] = cn.Node.ClientList[cn.Node.Id]
	// Run until the coordinator finds an alive node.
	for {
		successorId := cn.Node.findSuccessor(curId) // Finding the next node
		if successorId != -1 && successorId != cn.Node.CoordinatorId {
			client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000+successorId))
			if err != nil {
				done <- fmt.Errorf("[NODE-%d] Error connecting to successor %d: %s", cn.Node.Id, successorId, err)
				return
			}
			defer client.Close()

			var reply Message
			err = client.Call("ClientNode.DiscoverRingAndPropagate", msg, &reply)
			if err != nil {
				done <- fmt.Errorf("[NODE-%d] Error propagating ring discovery to node %d: %s", cn.Node.Id, successorId, err)
			} else {
				done <- nil
			}

			select {
			case err := <-done:
				if err != nil {
					fmt.Printf("%s\n", err)
					
					// If there is an error, remove the element from the client list and the ring structure
					cn.Node.Ring = deleteElement(cn.Node.Ring, successorId)
					delete(cn.Node.ClientList, successorId)
				} else {
					fmt.Printf("[NODE-%d] Ring Discovery propagated to node %d %v %v \n", cn.Node.Id, successorId, cn.Node.Ring, msg.Ring)
					return
				}
			case <-time.After(5 * time.Second):
				fmt.Printf("[NODE-%d] Timeout error. Could not communicate with node %d\n", cn.Node.Id, successorId)
			}
		} else { // Stop when we have reached the new node
			fmt.Printf("[NODE-%d] Ring discovery propagation completed.\n", cn.Node.Id)
			coordinator, err := rpc.Dial("tcp", LOCALHOST + "8000") // Communicating to the coordinator to begin announcement
			if err != nil {
				fmt.Printf("[NODE-%d] Error connecting to the coordinator. Begin Election process\n", cn.Node.Id)
			}
			defer coordinator.Close()

			fmt.Printf("[NODE-%d] Ring structure updated. New ring: %v\n", cn.Node.Id, msg.Ring)

			var reply Message
			err = coordinator.Call("CoordinatorNode.InitiateRingUpdate", msg, &reply)
			if err != nil {
				fmt.Printf("[NODE-%d] Error initiating ring update: %s\n", cn.Node.Id, err)
			} else {
				fmt.Printf("[NODE-%d] Ring update propagation initiated.\n", cn.Node.Id)
			}
			return
		}
	}
}

// Checks for sync message timeout
func (cn *ClientNode) CheckForTimeout() {
    for {
        cn.Node.Lock.Lock()  
        elapsed := time.Since(cn.LastUpdated)
        cn.Node.Lock.Unlock()

        // Check if more than 5 seconds have passed since the last update
        if elapsed > 5*time.Second {
            fmt.Printf("[NODE-%d] WARNING: Replica has not been synchronized for %v seconds. Last updated at %v. Initiate election process\n", cn.Node.Id, int(elapsed.Seconds()), cn.LastUpdated)
        }

		// Grace period for the coordinator to send sync messages
        time.Sleep(1 * time.Second)
    }
}

