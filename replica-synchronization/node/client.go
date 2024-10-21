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

func (cn *ClientNode) InvokeElection(msg *Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	fmt.Printf("[NODE-%d] Election initiated. Current coordinator is node-%d\n", cn.Node.Id, cn.Node.CoordinatorId)

	// If the current node is the coordinator, it will start the election process
	if cn.Node.Id == cn.Node.CoordinatorId {
		go cn.Node.ElectCoordinator()
	}

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

func (cn *ClientNode) InvokeSynchronization(msg *Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	cn.Node.LocalReplica = msg.Payload
	cn.LastUpdated = time.Now()

	fmt.Printf("[NODE-%d] Replica synchronized with the coordinator. Replica: '%v'. Ring structure %v\n", cn.Node.Id, msg.Payload, cn.Node.Ring)
	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

func (cn *ClientNode) DiscoverRing(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	curId := cn.Node.Id

	// Only to update the ring structure to include the new node.
	if cn.Node.Id != msg.NodeId {
		cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.findIndex(cn.Node.CoordinatorId), msg.NodeId)// Add the new node to the ring structure
	}

	msg.Ring = append(msg.Ring, cn.Node.Id)
	msg.ClientList[cn.Node.Id] = cn.Node.ClientList[cn.Node.Id]

	// Run until the coordinator finds an alive node.
	for {
		successorId := cn.Node.findSuccessor(curId) // Finding the next node
		curId = successorId

		if successorId != -1 && successorId != cn.Node.CoordinatorId {
			
			err := cn.propagateToSuccessor(successorId, msg, "ClientNode.DiscoverRing")
			if err != nil {
				fmt.Printf("%s\n", err)
				
				// If there is an error, remove the element from the client list and the ring structure
				cn.Node.Ring = deleteElement(cn.Node.Ring, successorId)
				delete(cn.Node.ClientList, successorId)
			} else {
				fmt.Printf("[NODE-%d] Ring Discovery propagated to node %d, discovered ring: %v \n", cn.Node.Id, successorId, msg.Ring)
				break
			}

		}else { // Stop when we have reached the new node
			fmt.Printf("[NODE-%d] Ring discovery propagation completed.\n", cn.Node.Id)
			coordinator, err := rpc.Dial("tcp", LOCALHOST + "8000") // Communicating to the coordinator to begin announcement
			if err != nil {
				fmt.Printf("[NODE-%d] Error connecting to the coordinator. Begin Election process\n", cn.Node.Id)
			}
			defer coordinator.Close()
	
			fmt.Printf("[NODE-%d] New ring structure discovered: %v\n", cn.Node.Id, msg.Ring)
	
			var reply Message
			err = coordinator.Call("CoordinatorNode.InitiateRingUpdate", msg, &reply)
			if err != nil {
				fmt.Printf("[NODE-%d] Error initiating ring update: %s\n", cn.Node.Id, err)
			} else {
				fmt.Printf("[NODE-%d] Ring update propagation initiated.\n", cn.Node.Id)
			}
			break
		}
	}
	
	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

func (cn *ClientNode) UpdateRing(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	cn.Node.ClientList = msg.ClientList
	cn.Node.Ring = msg.Ring
	cn.Node.CoordinatorId = msg.CoordinatorId

	// Propagate to the rest of the ring and update their ring structure
	successorId := cn.Node.findSuccessor(cn.Node.Id)

	if successorId != -1 && successorId != cn.Node.CoordinatorId {
		go func() {
			err := cn.propagateToSuccessor(successorId, msg, "ClientNode.UpdateRing")
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("[NODE-%d] Ring update propagated to node %d\n", cn.Node.Id, successorId)
			}
		}()
	} else {
		fmt.Printf("[NODE-%d] No successor found. Updating Ring structure is complete.\n", cn.Node.Id)
	}

	fmt.Printf("[NODE-%d] Ring structure updated. New ring: %v\n", cn.Node.Id, cn.Node.Ring)

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}

	return nil
}

func (cn *ClientNode) propagateToSuccessor(successorId int, msg Message, rpcCall string) error {
	client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000+successorId))
	if err != nil {
		return fmt.Errorf("[NODE-%d] Error connecting to successor %d: %s", cn.Node.Id, successorId, err)
	}
	defer client.Close()

	var reply Message
	err = client.Call(rpcCall, msg, &reply)
	if err != nil {
		return fmt.Errorf("[NODE-%d] Error propagating ring discovery to node %d: %s", cn.Node.Id, successorId, err)
	} else {
		return nil
	}
}

// Checks for sync message timeout
func (cn *ClientNode) CheckForTimeout() {
    for {
        cn.Node.Lock.Lock()  
        elapsed := time.Since(cn.LastUpdated)
        cn.Node.Lock.Unlock()

        // Check if more than 6 seconds have passed since the last update. 1 second grace period
        if elapsed > 6*time.Second {
            fmt.Printf("[NODE-%d] WARNING: Replica has not been synchronized for %v seconds. Last updated at %v. Initiate election process\n", cn.Node.Id, int(elapsed.Seconds()), cn.LastUpdated)
        }

		time.Sleep(1 * time.Second) // Reminds every second
    }
}

