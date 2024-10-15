package main

import (
	"fmt"
	"vector-clock/client"
	"vector-clock/server"
	"sync"
)

const (
	NumNodes = 10
)

func main() {

	clientChannels := make([]chan client.Message, NumNodes)  // creating a slice of channels for server to send messages to the clients
	serverChannels := make([]chan client.Message, NumNodes)  // creating a slice of channels for clients to send messages to the server

	for i := range clientChannels {
		clientChannels[i] = make(chan client.Message)
		serverChannels[i] = make(chan client.Message)
	}

	server := server.Server{make([]int, NumNodes + 1), clientChannels, serverChannels, sync.Mutex{}}

	go server.ReceiveMessage()
	for i := range clientChannels {
		client := client.Client{int(i), serverChannels[i], clientChannels[i], make([]int, NumNodes + 1)} // every client starts off with a logical clock of 0
		go client.SendMessage()
		go client.ReceiveMessage()
	}

	var input string
	fmt.Scanln(&input)
}
