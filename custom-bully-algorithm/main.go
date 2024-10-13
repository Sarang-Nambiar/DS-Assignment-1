package main

import (
	"custom-bully-algorithm/client"
	"custom-bully-algorithm/server"
	"fmt"
)


func main() {
	messages := make(chan string)
	clientChannels := make([] chan string, 10)
	
	for i := range clientChannels {
		clientChannels[i] = make(chan string)
	}
	
	server := server.Server{false, clientChannels}

	go server.ReceiveMessage(messages)
	for i := 0; i < 10; i++ {
		client := client.Client{uint8(i), clientChannels[i]}
		go client.SendMessage(messages)
		go client.ReceiveMessage()
	}

	var input string
	fmt.Scanln(&input)
}