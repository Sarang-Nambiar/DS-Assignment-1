package main

import (
	"fmt"
	"lamports-clock/client"
	"lamports-clock/server"
	"sync"
)

const(
	numNodes = 3
)

func main() {
	clientChannels := make([]chan client.Message, numNodes)  // creating a slice of channels for server to send messages to the clients
	serverChannels := make([]chan client.Message, numNodes)  // creating a slice of channels for clients to send messages to the server

	for i := range clientChannels {
		clientChannels[i] = make(chan client.Message)
		serverChannels[i] = make(chan client.Message)
	}

	server := server.Server{0, clientChannels, serverChannels, sync.Mutex{}} // every server starts off with a logical clock of 0

	go server.ReceiveMessage()
	for i := range clientChannels {
		client := client.Client{int(i), serverChannels[i], clientChannels[i], 0} // every client starts off with a logical clock of 0
		go client.SendMessage()
		go client.ReceiveMessage()
	}

	var input string
	fmt.Scanln(&input)
}
