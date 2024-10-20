package node

import (
	"fmt"
	"net/rpc"
	"slices"
	"strconv"
	"time"
)

type CoordinatorNode struct {
	Node *Node
}

func (cn *CoordinatorNode) RegisterNode(msg *Message, reply *Message) error {
	cn.Node.Lock.Lock()
	defer cn.Node.Lock.Unlock()

	// Initiate Ring discover and ring updating
	go cn.InitiateRingDiscovery(msg)

	*reply = Message{
		Type:    ACK,
		NodeId:  cn.Node.Id,
		Payload: cn.Node.LocalReplica,
	}
	return nil
}

// Starts from the coordinator node and then works its way around the ring to finish right before the new node, then it would start the new ring update process
// This function would be different for elections
// Fix the ACK messages after the ring discovery is complete.
// func (cn *CoordinatorNode) InitiateRingDiscovery(msg *Message) {
// 	curId := cn.Node.Id

// 	// Initialize an empty ring and clientlist
// 	newClientList := make(map[int]string)
// 	newRing := []int{}

// 	// Add the new node to the client list and the ring structure
// 	// Doing this so that the new node is counted in the discovery phase
// 	cn.Node.ClientList[msg.NodeId] = msg.ClientList[msg.NodeId]
// 	cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.findIndex(cn.Node.Id), msg.NodeId)

// 	newClientList[cn.Node.Id] = LOCALHOST + "8000"
// 	newRing = append(newRing, cn.Node.Id)

// 	// Run until the coordinator finds an alive node.
// 	for {
// 		successorId := cn.Node.findSuccessor(curId)
// 		curId = successorId

// 		msg := Message{
// 			Type:       msg.Type,
// 			NodeId:     msg.NodeId, // ID of the coordinator
// 			ClientList: newClientList,
// 			Ring:       newRing,
// 		}

// 		if successorId == -1 || successorId == cn.Node.Id {
// 			// If there is no successor, then assign the old ring to the new ring structure
// 			fmt.Printf("[COORDINATOR-%d] No successor found. Updating Ring structure is unnecessary. %v\n", cn.Node.Id, cn.Node.Ring)
// 			return
// 		}

// 		go func() {
// 			client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000+successorId))
// 			if err != nil {
// 				done <- fmt.Errorf("[COORDINATOR-%d] Error connecting to successor %d: %s", cn.Node.Id, successorId, err)
// 				return
// 			}
// 			defer client.Close()

// 			var reply Message
// 			err = client.Call("ClientNode.DiscoverRingAndPropagate", msg, &reply)
// 			if err != nil {
// 				done <- fmt.Errorf("[COORDINATOR-%d] Error propagating ring update to node %d: %s", cn.Node.Id, successorId, err)
// 				return
// 			} else {
// 				done <- nil
// 			}
// 		}()

// 		select {
// 		case err := <-done:
// 			if err != nil {
// 				fmt.Printf("%s\n", err)

// 				// If there is an error, remove the element from the client list and the ring structure
// 				cn.Node.Ring = deleteElement(cn.Node.Ring, successorId)
// 				delete(cn.Node.ClientList, successorId)
// 			} else {
// 				fmt.Printf("[COORDINATOR-%d] Ring discovery propagation initiated starting with node %d\n", cn.Node.Id, successorId)
// 				return // Check if this return belongs to the function above or the select statement
// 			}
// 		case <-time.After(5 * time.Second):
// 			fmt.Printf("[COORDINATOR-%d] Timeout error. Could not communicate with node %d\n", cn.Node.Id, successorId)
// 		}
// 	}
// }

func (cn *CoordinatorNode) InitiateRingDiscovery(msg *Message) {
	curId := cn.Node.Id
	done := make(chan error, 1)

	// Initialize an empty ring and clientlist
	newClientList := make(map[int]string)
	newRing := []int{}

	// Add the new node to the client list and the ring structure
	// Doing this so that the new node is counted in the discovery phase
	cn.Node.ClientList[msg.NodeId] = msg.ClientList[msg.NodeId]
	cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.findIndex(cn.Node.Id), msg.NodeId)

	newClientList[cn.Node.Id] = LOCALHOST + "8000"
	newRing = append(newRing, cn.Node.Id)
	
	// Run until the coordinator finds an alive node.
	for {
		successorId := cn.Node.findSuccessor(curId)
		curId = successorId

		msg := Message{
			Type:       msg.Type,
			NodeId:     msg.NodeId, // ID of the new node
			ClientList: newClientList,
			Ring:       newRing,
		}

		done <- cn.propagateToSuccessor(successorId, msg, "ClientNode.DiscoverRingAndPropagate")

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("%s\n", err)
	
				// If there is an error, remove the element from the client list and the ring structure
				cn.Node.Ring = deleteElement(cn.Node.Ring, successorId)
				delete(cn.Node.ClientList, successorId)
			} else {
				fmt.Printf("[COORDINATOR-%d] Ring discovery propagation initiated starting with node %d\n", cn.Node.Id, successorId)
				return
			}
		case <-time.After(5 * time.Second):
			fmt.Printf("[COORDINATOR-%d] Timeout error. Could not communicate with node %d\n", cn.Node.Id, successorId)
		}
	}
}

func (cn *CoordinatorNode) propagateToSuccessor(successorId int, msg Message, rpcCall string) error {

	if successorId == -1 || successorId == cn.Node.Id { // If the successor is not found or is the coordinator, then stop the ring update propagation
		fmt.Printf("[COORDINATOR-%d] No successor found. Updating Ring structure is unnecessary.\n", cn.Node.Id)
		return nil
	}


	client, err := rpc.Dial("tcp", LOCALHOST + strconv.Itoa(8000+successorId))
	if err != nil {
		return fmt.Errorf("[COORDINATOR-%d] Error connecting to successor %d: %s", cn.Node.Id, successorId, err)
	}
	defer client.Close()

	var reply Message
	err = client.Call(rpcCall, msg, &reply)
	if err != nil {
		return fmt.Errorf("[COORDINATOR-%d] Error propagating ring update to node %d: %s", cn.Node.Id, successorId, err)
	} 

	return nil
}

func (cn *CoordinatorNode) InitiateRingUpdate(msg Message, reply *Message) error {
	fmt.Printf("[COORDINATOR-%d] Initiating ring update. Updated ring structure %v\n", cn.Node.Id, msg.Ring)
	return nil
}

// // Issue with this is that if a node is not found, then even though the node which found
// func (cn *CoordinatorNode) propagateRingUpdate() {
// 	curId := cn.Node.Id
// 	done := make(chan error, 1)

// 	// Run until the coordinator finds an alive node.
// 	for {
// 		successorId := cn.Node.findSuccessor(curId)
// 		curId = successorId
// 		if successorId == -1 {
// 			fmt.Printf("[COORDINATOR-%d] No successor found. Updating Ring structure is unnecessary.\n", cn.Node.Id)
// 			return
// 		}
// 		msg := Message{
// 			Type:       DISCOVER,
// 			NodeId:     cn.Node.Id,
// 			ClientList: cn.Node.ClientList,
// 			Ring:       cn.Node.Ring,
// 		}

// 		go func() {
// 			client, err := rpc.Dial("tcp", cn.Node.ClientList[successorId])
// 			if err != nil {
// 				fmt.Printf("[COORDINATOR-%d] Error connecting to successor %d: %s\n", cn.Node.Id, successorId, err)
// 				return
// 			}
// 			defer client.Close()

// 			var reply Message
// 			err = client.Call("ClientNode.UpdateRingAndPropagate", msg, &reply)
// 			if err != nil {
// 				done <- fmt.Errorf("[COORDINATOR-%d] Error propagating ring update to node %d: %s\n", cn.Node.Id, successorId, err)
// 			} else {

// 			}
// 		}()

// 		select {
// 		case err := <-done:
// 			if err != nil {
// 				fmt.Printf("%s\n", err)
// 				deleteElement()
// 			}
// 			fmt.Printf("[COORDINATOR-%d] Ring update propagation initiated starting with node %d\n", cn.Node.Id, successorId)

// 		case <-time.After(5 * time.Second):

// 		}
// 	}
// }
