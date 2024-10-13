package main

import (
	"fmt"
	"sync"
	"custom-bully-algorithm/server"
	"custom-bully-algorithm/client"
)


func main() {
	var wg sync.WaitGroup
	messages := make(chan string)
	server := server.Server{
		drop: false,
	}
	go server.receiveMessage(messages)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		client := client.Client{uint8(i)}
		go client.sendMessage(messages)
	}
}