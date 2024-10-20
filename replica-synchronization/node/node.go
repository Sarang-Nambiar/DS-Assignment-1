package node

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Order of business
// 1. Whenever a node joins, we run the ring protocol to first do discovery, and then set the values of the client list and the ring structure in the next round of traversal
// 2. Implement a timeout feature if the synchronization doesn't happen in the next 5 seconds from the last updated time

type Node struct {
	Id            int
	LocalReplica  []int
	ClientList    map[int]string // Map over array because we can easily add or remove a node without indexing error
	Ring          []int
	IsCoordinator bool
	CoordinatorId int
	Lock          sync.Mutex
}

const (
	ACK      = "ACK" // Acknowledgement
	DISCOVER = "DISCOVER" // Discovery phase
	ANNOUNCE = "ANNOUNCE" // Announcement phase
	SYNC     = "SYNC" // Synchronizing replica
	UPDATE = "UPDATE" // Updating the ring structure(FOR NEWLY JOINED NODES ONLY)
	LOCALHOST = "127.0.0.1:"
)

func StartNode(node *Node) {
	cn := ClientNode{node, time.Now()}
	rpc.Register(&cn)

	listener, err := net.Listen("tcp", node.ClientList[node.Id])
	if err != nil {
		fmt.Printf("[NODE-%d] could not start listening: %s\n", node.Id, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("[NODE-%d] Node is running on %s\n", node.Id, node.ClientList[node.Id])

	go RegisterWithCoordinator(node)

	go cn.CheckForTimeout()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[NODE-%d] accept error: %s\n", node.Id, err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

// Check if you have to update the ring address value of the new coordinator nodee
func StartCoordinator(node *Node) {
	rpc.Register(&CoordinatorNode{node})
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Error starting coordinator:", node.Id), err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Coordinator is running on %s", node.Id, node.ClientList[node.Id]))

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Error listening to accepting incoming connections", node.Id))
		}

		go rpc.ServeConn(conn)
	}
}

// Double check if you need to initiate the ring join process when a node joins the network
func RegisterWithCoordinator(node *Node) {
	client, err := rpc.Dial("tcp", ":8000")

	if err != nil {
		fmt.Printf("[NODE-%d] Error connecting to coordinator. Please check if there is a running coordinator: %s\n", node.Id, err)
		return
	}
	defer client.Close()

	var request Message = Message{
		Type:       DISCOVER,
		NodeId:     node.Id,
		ClientList: node.ClientList,
		Ring:       node.Ring,
	}
	var reply Message
	err = client.Call("CoordinatorNode.RegisterNode", request, &reply)
	if err != nil {
		fmt.Printf("Error registering with coordinator: %s\n", err)
		return
	}

	if reply.Type == ACK {
		node.LocalReplica = reply.Payload
		node.CoordinatorId = reply.NodeId
	}

	fmt.Printf("[NODE-%d] Node has been registered with the coordinator.\n", node.Id)
}

func (n *Node) SynchronizeReplica() {
	for {
		fmt.Printf("[COORDINATOR-%d] Replica synchronization has begun, Replica: '%v'\n", n.CoordinatorId, n.LocalReplica)
		for i, v := range n.ClientList {
			if i != n.CoordinatorId {
				client, err := rpc.Dial("tcp", v)
				if err != nil {
					fmt.Printf("[COORDINATOR-%d] Error occurred while creating a connection between coordinator and node-%d: %s\n", n.CoordinatorId, i, err)
					continue
				}

				var reply Message
				var msg Message = Message{
					Type:    SYNC,
					NodeId:  n.CoordinatorId,
					Payload: n.LocalReplica,
				}
				err = client.Call("ClientNode.InvokeSynchronization", msg, &reply)

				if err != nil {
					fmt.Printf("[COORDINATOR-%d] Error occurred while receiving a response from the client node-%d: %s\n", n.CoordinatorId, n.Id, err)
					continue
				}

				if reply.Type == ACK {
					fmt.Printf("[COORDINATOR-%d] Replica successfully synchronized with client node %d\n", n.CoordinatorId, reply.NodeId)
				}
			}
		}

		time.Sleep(5 * time.Second) // Call synchronization every 5 seconds
	}
}

// Utility functions

func (n *Node) findSuccessor(id int) int {
	if len(n.Ring) <= 1 {
		return -1
	}

	for i := range n.Ring {
		if n.Ring[i] == id {
			if (i + 1) < len(n.Ring) {
				return n.Ring[i+1]
			}
			// If node is the last element in the Ring, then return the first
			return n.Ring[0]
		}
	}
	return -1 // If the node is not found in the Ring
}

func (n *Node) findIndex(element int) int {
	for i, v := range n.Ring {
		if v == element {
			return i
		}
	}
	return -1
}

func contains(nodeList map[int]string, i int) bool {
	if _, ok := nodeList[i]; ok {
		return true
	}
	return false
}

func GetUniqueId(nodeList map[int]string) int {
	for i := 1; ; i++ {
		if !contains(nodeList, i) {
			return i
		}
	}
}

func deleteElement(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}
