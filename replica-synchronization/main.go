package main

import (
	"fmt"
	"replica-synchronization/node"
	"sync"
)

const (
	numNodes = 5
)

func main(){
	// Create a 2D array of all the channels for communication between every other node.
	nodeChannels := make([]chan node.Message, numNodes - 1)
	ring := make([]int, numNodes)

	for j := range nodeChannels{
		nodeChannels[j] = make(chan node.Message)
		ring[j] = j
	}

	// define the nodes first
	for i := 0; i < numNodes - 1; i++ {
		// Initializing node with id numNodes - 1 as the coordinator
		n := node.Node{i, "", ring, false, numNodes - 1, nodeChannels, sync.Mutex{}}
		go n.ReceiveMessage()
	}
	
	// Initializing coordinator
	coordinator := node.Node{numNodes - 1, "This is a synchronized document", ring, true, numNodes - 1, nodeChannels, sync.Mutex{}}
	go coordinator.SendMessage()
	go coordinator.ReceiveMessage()

	var input string
	fmt.Scanln(&input)
}