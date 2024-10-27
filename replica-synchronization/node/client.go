package node

import (
	"fmt"
	"math/rand/v2"
	"net"
	"net/rpc"
	"slices"
	"strconv"
	"strings"
	"time"
)

// PROBLEMS:
// 1. The Joining node doesn't know who the coordinator is.
// Potential solution: store which node is the coordinator in the nodes-list.json file or another file.
type ClientNode struct {
	Node        *Node
	LastUpdated time.Time
	Listener   net.Listener // Client node listener to close the connection when elected as coordinator
	isCoordinator bool
}

// Invokes the discovery phase of the ring election process
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
	printWithDelay("[NODE-%d] Initiating the discovery phase of the ring election.\n", cn.Node.Id)
	err := cn.DiscoverRing(discoverMsg, &discoverReply)
	if err != nil {
		fmt.Printf("[NODE-%d] Error in DiscoverRing: %v\n", cn.Node.Id, err)
	}

	time.Sleep(1 * time.Second) // Wait for the discovery phase to complete

}

// Invoke synchronization of the replica with the coordinator.
func (cn *ClientNode) InvokeSynchronization(msg *Message, reply *Message) error {

	cn.Node.Lock.Lock()
	cn.Node.LocalReplica = msg.Payload
	cn.LastUpdated = time.Now()
	cn.Node.Lock.Unlock()

	fmt.Printf("[NODE-%d] Replica synchronized with the coordinator. Replica: '%v'. Ring structure %v\n", cn.Node.Id, msg.Payload, cn.Node.Ring)
	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}

	// Modify the replica randomly after synchronization to simulate real world scenarios
	go cn.modifyReplica()

	return nil
}

// DISCOVERY PHASE
// Function to discover the ring structure and the new coordinator
func (cn *ClientNode) DiscoverRing(msg Message, reply *Message) error {
	curId := cn.Node.Id

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

		if successorId != -1 && successorId != msg.CoordinatorId {
			err := cn.propagateToSuccessor(successorId, msg, "ClientNode.DiscoverRing")
			if err != nil {
				fmt.Printf("%s\n", err)
				if strings.Contains(err.Error(), "coordinator") {
					// If the successor node has been promoted to coordinator, then break out of the loop since the election is already complete.
					fmt.Printf("[NODE-%d] Ring Discovery propagation completed since a coordinator already exists.\n", cn.Node.Id)
					break
				}
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

// ANNOUNCEMENT PHASE
// Function to update the ring with the new ring structure and the coordinator
func (cn *ClientNode) UpdateRing(msg Message, reply *Message) error {
	curId := cn.Node.Id
	
	cn.Node.Lock.Lock()
	cn.Node.ClientList = msg.ClientList
	cn.Node.Ring = msg.Ring
	cn.Node.CoordinatorId = msg.CoordinatorId
	cn.LastUpdated = time.Now()
	cn.Node.Lock.Unlock()

	// Propagate to the rest of the ring and update their ring structure
	for {
		successorId := cn.Node.findSuccessor(curId)
		curId = successorId

		if successorId != -1 && successorId != msg.CoordinatorId {
			err := cn.propagateToSuccessor(successorId, msg, "ClientNode.UpdateRing")
			if err != nil {
				fmt.Printf("%s\n", err)
				if strings.Contains(err.Error(), "coordinator") {
					fmt.Printf("[NODE-%d] Ring update propagation completed since a coordinator already exists.\n", cn.Node.Id)
					break
				}
			} else {
				fmt.Printf("[NODE-%d] Ring update propagated to node %d. New coordinator is node %d\n", cn.Node.Id, successorId, msg.CoordinatorId)
				break
			}
		} else {
			fmt.Printf("[NODE-%d] No successor found. Updating Ring structure is complete.\n", cn.Node.Id)
			break
		}
	}
	

	fmt.Printf("[NODE-%d] Ring structure updated. New ring: %v\n", cn.Node.Id, cn.Node.Ring)

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}

	return nil
}

// Function to propagate through the ring
func (cn *ClientNode) propagateToSuccessor(successorId int, msg Message, rpcCall string) error {
	client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000+successorId))
	if err != nil {
		return fmt.Errorf("[NODE-%d] Error connecting to successor %d: %s", cn.Node.Id, successorId, err)
	}
	defer client.Close()

	var reply Message
	err = client.Call(rpcCall, msg, &reply)
	if err != nil {
		// This is done because if the node/coordinator fails, it is not reachable through rpc so this condition is not an error
		// Error occurs when accessing rpc methods of the node/coordinator
		fmt.Printf("[NODE-%d] Error propagating ring discovery to node %d. This is probably because the node has been promoted to coordinator. Error: %s", cn.Node.Id, successorId, err)
		return nil
	} else {
		return nil
	}
}

// Function to transition the elected ClientNode to a CoordinatorNode
func (cn *ClientNode) BecomeCoordinator(msg Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	if cn.isCoordinator {
		fmt.Printf("[NODE-%d] Already transitioned to a coordinator\n", cn.Node.Id)
		return nil
	}

	cn.isCoordinator = true
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
	
	fmt.Printf("[COORDINATOR-%d] Successfully transitioned to coordinator role\n", cn.Node.Id)
	return nil
}

// Function to handle the ring completion
func (cn *ClientNode) handleRingCompletion(msg Message) error {
    // If new node, then update the ring structure with the new node
	if msg.Type == NDISCOVER {
        return cn.handleNewNodeRingUpdate(msg)
    }
    return cn.handleElectionRingUpdate(msg)
}

// Function to handle the new node ring update
func (cn *ClientNode) handleNewNodeRingUpdate(msg Message) error {
	fmt.Printf("[NODE-%d] Initiating ring update for the new node. New Ring structure: %v\n", cn.Node.Id, msg.Ring)
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

// Function to handle the election ring update
func (cn *ClientNode) handleElectionRingUpdate(msg Message) error {
	fmt.Printf("[NODE-%d] Initiating the announcement phase of the election. Newly elected coordinator is node %d\n", cn.Node.Id, msg.CoordinatorId)
	// time.Sleep(5 * time.Second) 
    client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000 + msg.CoordinatorId))
    if err != nil {
        return fmt.Errorf("error connecting to new coordinator: %v", err)
    }
    defer client.Close()

	time.Sleep(5 * time.Second) // Uncomment for part 2.3 (a) and part 2.3 (b) to kill a node(coordinator or client) before the new coordinator ID is circulated through the ring.
    var reply Message
    if err := client.Call("ClientNode.UpdateRing", msg, &reply); err != nil {
        return fmt.Errorf("error updating ring: %v", err)
    }

	fmt.Printf("[NODE-%d] Ring update propagated to the new coordinator. New coordinator is node %d\n", cn.Node.Id, msg.CoordinatorId)

	// time.Sleep(5 * time.Second) // Uncomment for part 2.3 (a) right before the newly elected node becomes the coordinator.
    if err := client.Call("ClientNode.BecomeCoordinator", msg, &reply); err != nil {
        return fmt.Errorf("error converting to coordinator: %v", err)
    }

    fmt.Printf("[NODE-%d] Coordinator elected.\n", cn.Node.Id)
    return nil
}

// Checks for synchronization message timeout
func (cn *ClientNode) CheckForTimeout() {
    for {
		// If the node is the coordinator, then there is no need to check for elections
		if cn.isCoordinator {
			break
		}
        cn.Node.Lock.Lock()  
        elapsed := time.Since(cn.LastUpdated)
        cn.Node.Lock.Unlock()

        // Check if more than 6 seconds have passed since the last update. 1 second grace period
		if elapsed > time.Duration(6 + cn.Node.Id) * time.Second {
            fmt.Printf("[NODE-%d] WARNING: Replica has not been synchronized for %v seconds.\n", cn.Node.Id, int(elapsed.Seconds()))
			// time.Sleep(5 * time.Second) // Uncomment this for part 2.2 to simulate simultaneous elections
			go cn.InvokeElection()
        }

		time.Sleep(1 * time.Second) // Reminds every second
    }
}

// Modifies the replica randomly
func (cn *ClientNode) modifyReplica() {
	time.Sleep(3 * time.Second)
	randIndex := rand.IntN(10) // Generates a number from 0 to 9

	randNum := rand.IntN(100) // Generates a number from 0 to 99

	cn.Node.Lock.Lock()
	cn.Node.LocalReplica[randIndex] = randNum
	fmt.Printf("[NODE-%d] Replica modified. New replica: '%v'\n", cn.Node.Id, cn.Node.LocalReplica)
	cn.Node.Lock.Unlock()
}

func printWithDelay(format string, a ...interface{}) {
	time.Sleep(1 * time.Second)
	fmt.Printf(format, a...)
}
