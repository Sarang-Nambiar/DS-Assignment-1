package node

import (
	"fmt"
	"net"
	"net/rpc"
	"slices"
	"strconv"
	"time"
)


// PROBLEMS:
// 1. The Joining node doesn't know who the coordinator is.
// Potential solution: store which node is the coordinator in the nodes-list.json file or another file.
type ClientNode struct {
	Node        *Node
	LastUpdated time.Time
	Listener   net.Listener // Client node listener to close the connection when elected as coordinator
}

func (cn *ClientNode) InvokeElection() {
	fmt.Printf("[NODE-%d] Election initiated\n", cn.Node.Id)

	discoverMsg := Message{
		Type:       DISCOVER,    
		NodeId:     cn.Node.Id,       
		Ring:       []int{},          
		ClientList: make(map[int]string), 
		CoordinatorId: cn.Node.Id, // Initialize the current node id the coordinator
	}

	var discoverReply Message
	err := cn.DiscoverRing(discoverMsg, &discoverReply)
	if err != nil {
		fmt.Printf("[NODE-%d] Error in DiscoverRing: %v\n", cn.Node.Id, err)
	}

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
	curId := cn.Node.Id
	removeNodeId := -1

	// Only to update the ring structure to include the new node.
	cn.Node.Lock.Lock()
	if msg.Type == NDISCOVER && cn.Node.Id != msg.NodeId {
		cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.findIndex(cn.Node.CoordinatorId), msg.NodeId)// Add the new node to the ring structure
	} else {
		if msg.Type == DISCOVER && cn.Node.Id > msg.CoordinatorId {
			// update coordinator id to the max(curNodeId, msg.CoordinatorId)
			// Reset the ring and client list
			msg.CoordinatorId = cn.Node.Id
			msg.Ring = []int{}
			msg.ClientList = map[int]string{}
		}
	}

	// Updating the new ring and client list
	msg.Ring = append(msg.Ring, cn.Node.Id)
	msg.ClientList[cn.Node.Id] = cn.Node.ClientList[cn.Node.Id]
	cn.LastUpdated = time.Now()
	cn.Node.Lock.Unlock()

	// Run until the coordinator finds an alive node.
	for {
		// Finding first alive successor
		successorId := cn.Node.findSuccessor(curId)
		curId = successorId

		fmt.Printf("Remove node is %d and curId: %d\n", removeNodeId, curId)
		if removeNodeId == curId {
			// If there is an error, remove the element from the client list and the ring structure
			cn.Node.Lock.Lock()
			cn.Node.Ring = deleteElement(cn.Node.Ring, removeNodeId)
			delete(cn.Node.ClientList, removeNodeId)
			removeNodeId = -1
			cn.Node.Lock.Unlock()
		}

		fmt.Printf("[NODE-%d] Successor ID: %d. Current Ring: %v, Message Ring: %v\n", cn.Node.Id, successorId, cn.Node.Ring, msg.Ring)

		if successorId != -1 && successorId != msg.CoordinatorId {
			err := cn.propagateToSuccessor(successorId, msg, "ClientNode.DiscoverRing")
			if err != nil {
				fmt.Printf("%s\n", err)
				removeNodeId = successorId
			} else {
				fmt.Printf("[NODE-%d] Ring Discovery propagated to node %d, discovered ring: %v \n", cn.Node.Id, successorId, msg.Ring)
				break
			}
		} else {
			fmt.Printf("[NODE-%d] Ring discovery propagation completed. New ring structure discovered: %v\n", cn.Node.Id, msg.Ring)
			if err := cn.handleRingCompletion(msg); err != nil {
				fmt.Printf("[NODE-%d] Error handling ring completion: %v\n", cn.Node.Id, err)
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


func (cn *ClientNode) handleRingCompletion(msg Message) error {
    if msg.Type == NDISCOVER {
        return cn.handleNewNodeRingUpdate(msg)
    }
    return cn.handleElectionRingUpdate(msg)
}

func (cn *ClientNode) handleNewNodeRingUpdate(msg Message) error {
    coordinator, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000 + msg.CoordinatorId))
    if err != nil {
        return fmt.Errorf("error connecting to coordinator: %v", err)
    }
    defer coordinator.Close()

    var reply Message
    err = coordinator.Call("CoordinatorNode.InitiateRingUpdate", msg, &reply)
    if err != nil {
        return fmt.Errorf("error initiating ring update: %v", err)
    }
    
    fmt.Printf("[NODE-%d] Ring update propagation initiated.\n", cn.Node.Id)
    return nil
}

func (cn *ClientNode) handleElectionRingUpdate(msg Message) error {
    client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000 + msg.CoordinatorId))
    if err != nil {
        return fmt.Errorf("error connecting to new coordinator: %v", err)
    }
    defer client.Close()

    var reply Message
    if err := client.Call("ClientNode.UpdateRing", msg, &reply); err != nil {
        return fmt.Errorf("error updating ring: %v", err)
    }

    if err := client.Call("ClientNode.BecomeCoordinator", msg, &reply); err != nil {
        return fmt.Errorf("error converting to coordinator: %v", err)
    }

    fmt.Printf("[NODE-%d] Coordinator elected.\n", cn.Node.Id)
    return nil
}

func (cn *ClientNode) UpdateRing(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	cn.Node.ClientList = msg.ClientList
	cn.Node.Ring = msg.Ring
	cn.Node.CoordinatorId = msg.CoordinatorId
	cn.Node.Lock.Unlock()

	// Propagate to the rest of the ring and update their ring structure
	successorId := cn.Node.findSuccessor(cn.Node.Id)

	if successorId != -1 && successorId != msg.CoordinatorId {
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

func (cn *ClientNode) BecomeCoordinator(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	cn.Node.CoordinatorId = cn.Node.Id

	coordinator := CoordinatorNode{Node: cn.Node}

	RPCServer := rpc.NewServer()

	if err := RPCServer.Register(&coordinator); err != nil {
		return fmt.Errorf("[NODE-%d] Error registering coordinator: %v", cn.Node.Id, err)
	}

	switchComplete := make(chan struct{})

    go func() {
        close(switchComplete)

        for {
            conn, err := cn.Listener.Accept()
            if err != nil {
                fmt.Printf("Accept error: %v\n", err)
                continue
            }
            go RPCServer.ServeConn(conn)
        }
    }()

	<-switchComplete // Wait for the listener to start
	
	// Begin Synchronization
	go coordinator.SynchronizeReplica()

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	
	fmt.Printf("[NODE-%d] Successfully transitioned to coordinator role\n", cn.Node.Id)
	return nil
}

// Checks for sync message timeout
func (cn *ClientNode) CheckForTimeout() {
    for {
		// If the node is the coordinator, then there is no need to check for elections
		if cn.Node.Id == cn.Node.CoordinatorId {
			break
		}
        cn.Node.Lock.Lock()  
        elapsed := time.Since(cn.LastUpdated)
        cn.Node.Lock.Unlock()

        // Check if more than 6 seconds have passed since the last update. 1 second grace period
		if elapsed > time.Duration(6+ cn.Node.Id) * time.Second {
            fmt.Printf("[NODE-%d] WARNING: Replica has not been synchronized for %v seconds.\n", cn.Node.Id, int(elapsed.Seconds()))
			go cn.InvokeElection()
        }

		time.Sleep(1 * time.Second) // Reminds every second
    }
}

func (cn *ClientNode) StopClientNode() error {
	if cn.Listener != nil {
		fmt.Printf("[NODE-%d] Stopping ClientNode RPC server.\n", cn.Node.Id)
		return cn.Listener.Close()
	}

	time.Sleep(500 * time.Millisecond) // Some time for the listener to close
	return nil
}

