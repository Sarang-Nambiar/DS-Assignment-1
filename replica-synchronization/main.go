package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"replica-synchronization/node"
	"strconv"
	"sync"
	"syscall"
)

func main() {
	n := node.Node{Lock: sync.Mutex{}, ClientList: make(map[int]string), Ring: make([]int, 0)}

	nodesList := readNodesList()
	maxId := -1 // Initialize the maxId to -1

	// If there are no nodes running in the network, assign this node to be the coordinator by default
	if len(nodesList) == 0 {
		n.Id = 0
		n.LocalReplica = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		n.Ring = append(n.Ring, n.Id)
		n.ClientList[n.Id] = node.LOCALHOST + "8000"
		go node.StartCoordinator(&n)
	} else {
		n.Id = node.GetUniqueId(nodesList)
		n.ClientList = nodesList
		n.ClientList[n.Id] = node.LOCALHOST + strconv.Itoa(8000+n.Id)
		
		// Check if the initial coordinator is still present in the network
		if _, ok := nodesList[0]; ok {
			n.CoordinatorId = 0
		} else {
			n.CoordinatorId = maxId
		}
	}

	jsonData, err := json.Marshal(n.ClientList)

	err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)

	if err != nil {
		fmt.Println("Error occurred while marshalling the Ring back into nodes-list.json: ", err)
	}

	if n.CoordinatorId != n.Id {
		go node.StartNode(&n)
	}

	// Handle shutting down case
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Shutting down...")

		// Remove the node from the list
		nodesList = readNodesList()

		delete(nodesList, n.Id)

		jsonData, err := json.Marshal(nodesList)
		err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("Error occurred while updating nodes-list.json: ", err)
		}
		os.Exit(0)
	}()

	select {} // Blocking the main function from exiting immediately
}

func readNodesList() map[int]string {
	jsonFile, err := os.Open("nodes-list.json")
	if err != nil {
		fmt.Println("Error opening nodes-list.json file:", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var nodesList map[int]string

	json.Unmarshal(byteValue, &nodesList) // Puts the byte value into the nodesList map

	return nodesList
}
