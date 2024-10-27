package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	nodes map[int]string
	CoordinatorId int
}

type Node struct {
	Id            int
	LocalReplica  []int
	ClientList    map[int]string // Map over array because we can easily add or remove a node without indexing error
	Ring          []int
	CoordinatorId int
	Lock          sync.Mutex
}

const (
	ACK      = "ACK" // Acknowledgement
	DISCOVER = "DISCOVER" // Discovery phase during election
	ANNOUNCE = "ANNOUNCE" // Announcement phase during election
	SYNC     = "SYNC" // Synchronizing replica
	NDISCOVER = "NDISCOVER" // New node discovery
	LOCALHOST = "127.0.0.1:"
)

// Function to start a ClientNode
func StartNode(node *Node) {
	cn := ClientNode{
		Node: node, 
		LastUpdated: time.Now(),
	}

	rpc.Register(&cn)

	listener, err := net.Listen("tcp", node.ClientList[node.Id])
	cn.Listener = listener
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

// Function to start the coordinator node when there are no nodes in the network.
// Default coordinator is set to node with id 0
func StartCoordinator(node *Node) {
	node.CoordinatorId = node.Id

	cn := CoordinatorNode{node}
	rpc.Register(&cn)
	listener, err := net.Listen("tcp", node.ClientList[node.Id])
	if err != nil {
		fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Error starting coordinator:", node.Id), err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("[COORDINATOR-%d] Coordinator is running on %s\n", node.Id, node.ClientList[node.Id])

	// Begin Synchronization
	go cn.SynchronizeReplica()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("[COORDINATOR-%d] Error listening to accepting incoming connections\n", node.Id)
		}

		go rpc.ServeConn(conn)
	}
}

// Registering a new ClientNode with the Coordinator
func RegisterWithCoordinator(node *Node) {
	client, err := rpc.Dial("tcp", node.ClientList[node.CoordinatorId])

	if err != nil {
		fmt.Printf("[NODE-%d] Error connecting to coordinator. Please check if there is a running coordinator: %s\n", node.Id, err)
		return
	}
	defer client.Close()

	var request Message = Message{
		Type:       NDISCOVER,
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

// Function to find the index of an element in the Ring
func (n *Node) FindIndex(element int) int {
	for i, v := range n.Ring {
		if v == element {
			return i
		}
	}
	return -1
}

// Function to check if an element is inside a map
func contains(nodeList map[int]string, i int) bool {
	if _, ok := nodeList[i]; ok {
		return true
	}
	return false
}

// Getting a unique id for a newly joined node
func GetUniqueId(nodeList map[int]string) int {
	for i := 1; ; i++ {
		if !contains(nodeList, i) {
			return i
		}
	}
}

// Function to delete an element from a slice
func deleteElement(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}

// Function to read the nodes-list.json file and return the node ids along with their addresses
func ReadNodesList() map[string]interface{} {
	jsonFile, err := os.Open("nodes-list.json")
	if err != nil {
		fmt.Println("Error opening nodes-list.json file:", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var nodesList map[string]interface{}

	json.Unmarshal(byteValue, &nodesList) // Puts the byte value into the nodesList map

	if nodes, ok := nodesList["nodes"].(map[string]interface{}); ok {
		nodesList["nodes"] = ConvertMapStringToInt(nodes)
	}

	return nodesList
}

// Converting map[string]string to map[int]string
func ConvertMapStringToInt(input map[string]interface{}) map[int]string {
    result := make(map[int]string)
    for k, v := range input {
        if id, err := strconv.Atoi(k); err == nil {
			result[id] = v.(string)
        }
    }
    return result
}
