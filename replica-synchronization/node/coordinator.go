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

// Function to register a new ClientNode with the CoordinatorNode
func (cn *CoordinatorNode) RegisterNode(msg *Message, reply *Message) error {

	// Initiate Ring discover and ring updating
	go cn.InitiateRingDiscovery(msg)

	*reply = Message{
		Type:    ACK,
		NodeId:  cn.Node.Id,
	}
	return nil
}

// Function to synchronize the replica with the rest of the nodes in the network
func (cn *CoordinatorNode) SynchronizeReplica() {
	for {
		fmt.Printf("[COORDINATOR-%d] Replica synchronization has begun, Replica: '%v'. Ring: %v\n", cn.Node.CoordinatorId, cn.Node.LocalReplica, cn.Node.Ring)

		if len(cn.Node.ClientList) == 1 {
			fmt.Printf("[COORDINATOR-%d] No other nodes to synchronize with.\n", cn.Node.Id)
			time.Sleep(5 * time.Second)
			continue
		}

		for i, v := range cn.Node.ClientList {
			if i != cn.Node.CoordinatorId {
				client, err := rpc.Dial("tcp", v)
				if err != nil {
					fmt.Printf("[COORDINATOR-%d] Error occurred while creating a connection between coordinator and node-%d: %s\n", cn.Node.Id, i, err)
					continue
				}

				var reply Message
				var msg Message = Message{
					Type:    SYNC,
					NodeId:  cn.Node.Id,
					Payload: cn.Node.LocalReplica,
				}
				err = client.Call("ClientNode.InvokeSynchronization", msg, &reply)

				if err != nil {
					fmt.Printf("[COORDINATOR-%d] Error occurred while receiving a response from the client node-%d: %s\n", cn.Node.Id, cn.Node.Id, err)
					continue
				}

				if reply.Type == ACK {
					fmt.Printf("[COORDINATOR-%d] Replica successfully synchronized with client node %d\n", cn.Node.Id, reply.NodeId)
				}
			}
		}

		time.Sleep(5 * time.Second) // Call synchronization every 5 seconds
	}
}

// FOR NEW NODE ADDITION
// Function to initiate the ring discovery propagation within the client nodes.
func (cn *CoordinatorNode) InitiateRingDiscovery(msg *Message) {
	curId := cn.Node.Id

	// Initialize an empty ring and clientlist
	newClientList := make(map[int]string)
	newRing := []int{}

	// Add the new node to the client list and the ring structure
	// Doing this so that the new node is counted in the discovery phase
	cn.Node.Lock.Lock()
	cn.Node.ClientList[msg.NodeId] = msg.ClientList[msg.NodeId]
	cn.Node.Ring = slices.Insert(cn.Node.Ring, cn.Node.FindIndex(cn.Node.Id), msg.NodeId)

	newClientList[cn.Node.Id] = cn.Node.ClientList[cn.Node.CoordinatorId]
	newRing = append(newRing, cn.Node.Id)
	cn.Node.Lock.Unlock()
	
	// Run until the coordinator finds an alive node.
	for {
		successorId := cn.Node.findSuccessor(curId)
		curId = successorId

		msg := Message{
			Type:       msg.Type,
			NodeId:     msg.NodeId, // ID of the new node
			ClientList: newClientList,
			Ring:       newRing,
			CoordinatorId: cn.Node.Id,
		}

		if successorId == -1 || successorId == cn.Node.Id { // If the successor is not found or is the coordinator, then stop the ring update propagation
			fmt.Printf("[COORDINATOR-%d] No successor found. Updating Ring structure is unnecessary.\n", cn.Node.Id)
			return
		}
		
		err := cn.propagateToSuccessor(successorId, msg, "ClientNode.DiscoverRing")
		if err != nil {
			fmt.Printf("%s\n", err)

			// If there is an error, remove the element from the client list and the ring structure
			cn.Node.Lock.Lock()
			cn.Node.Ring = deleteElement(cn.Node.Ring, successorId)
			delete(cn.Node.ClientList, successorId)
			cn.Node.Lock.Unlock()
		} else {
			fmt.Printf("[COORDINATOR-%d] Ring discovery propagation initiated starting with node %d\n", cn.Node.Id, successorId)
			// Break out of the loop after initiating the ring discovery propagation successfully
			break
		}
	}
}

// FOR NEW NODE ADDITION
// Function to initiate the ring update propagation within the client nodes.
func (cn *CoordinatorNode) InitiateRingUpdate(msg Message, reply *Message) error {
	fmt.Printf("[COORDINATOR-%d] Ring update propagation initiated.\n", cn.Node.Id)

	// Update the ring structure
	cn.Node.Lock.Lock()
	cn.Node.Ring = msg.Ring
	cn.Node.ClientList = msg.ClientList
	msg.CoordinatorId = cn.Node.Id
	cn.Node.Lock.Unlock()

	fmt.Printf("[COORDINATOR-%d] Ring structure updated. New ring from msg: %v\n", cn.Node.Id, msg.Ring)

	// Propagate to the rest of the ring and update their ring structure
	successorId := cn.Node.findSuccessor(cn.Node.Id)

	if successorId == -1 || successorId == cn.Node.Id { // If the successor is not found or is the coordinator, then stop the ring update propagation
		fmt.Printf("[COORDINATOR-%d] No successor found. Updating Ring structure is unnecessary.\n", cn.Node.Id)
		return nil
	}

	go func () {
		err := cn.propagateToSuccessor(successorId, msg, "ClientNode.UpdateRing")
		if err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("[COORDINATOR-%d] Ring update propagated to node %d\n", cn.Node.Id, successorId)
		}
	}()

	*reply = Message{
		Type:   ACK,
		NodeId: cn.Node.Id,
	}
	return nil
}

// Function to propagate through the ring
func (cn *CoordinatorNode) propagateToSuccessor(successorId int, msg Message, rpcCall string) error {

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