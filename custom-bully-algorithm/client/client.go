package client

import (
	"fmt"
	"time"
)

type Client struct{
	id uint8 // unsigned integer of 8 bits
}

// send message function to server
func (c *Client) sendMessage(messages chan string) {
	for{
		time.Sleep(2 * time.Second) // every two seconds, a message is sent
		fmt.Println("Sending message to server from client")
	}
}

func (c *Client) receiveMessage(channel chan string){
	fmt.Println("Receiving message from server")
}