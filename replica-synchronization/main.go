package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"replica-synchronization/node"
	"slices"
	"strconv"
	"sync"
	"syscall"
)

func main() {
	// Create a new node instance
	n := node.Node{
		Lock: sync.Mutex{}, 
		ClientList: make(map[int]string), 
		Ring: make([]int, 0),
	}

	nodesList := node.ReadNodesList()

	nodes := make(map[int]string)
	if nodesList["nodes"] != nil {
		nodes = nodesList["nodes"].(map[int]string)
	}

	// If there are no nodes running in the network, assign this node to be the coordinator by default
	if len(nodes) == 0 {
		n.Id = 0
		n.LocalReplica = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		n.Ring = append(n.Ring, n.Id)
		n.ClientList[n.Id] = node.LOCALHOST + "8000"
		nodesList["nodes"] = n.ClientList
		nodesList["coordinator"] = strconv.Itoa(n.Id)
		go node.StartCoordinator(&n)
	} else {
		n.Id = node.GetUniqueId(nodes) // asserting the type to map[int]string
		n.ClientList = nodes
		for id := range n.ClientList {
			n.Ring = append(n.Ring, id)
		}
		n.ClientList[n.Id] = node.LOCALHOST + strconv.Itoa(8000+n.Id)

		// Inserting the new node right before the coordinator in the ring
		coordinatorId, err := strconv.Atoi(nodesList["coordinator"].(string))

		if err != nil {
			fmt.Println("Error occurred while converting coordinatorId to int: ", err)
		}

		coordinatorIndex := n.FindIndex(coordinatorId)
		n.CoordinatorId = coordinatorId
		n.Ring = slices.Insert(n.Ring, coordinatorIndex, n.Id) 
	}

	nodesList["nodes"] = n.ClientList
	jsonData, err := json.Marshal(nodesList)

	err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)

	if err != nil {
		fmt.Println("Error occurred while marshalling the Ring back into nodes-list.json: ", err)
	}

	if n.CoordinatorId != n.Id {
		go node.StartNode(&n)
	}

	// Handling when the node fails or is shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Shutting down...")

		// Remove the node from the list
		nodesList = node.ReadNodesList()

		delete(nodesList["nodes"].(map[int]string), n.Id)

		jsonData, err := json.Marshal(nodesList)
		err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("Error occurred while updating nodes-list.json: ", err)
		}
		os.Exit(0)
	}()

	select {} // Blocking the main function from exiting immediately
}
