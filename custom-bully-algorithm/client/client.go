package client

import (
	"fmt"
	"time"
)

type Client struct{
	Id uint8 // unsigned integer of 8 bits
	SubChannel chan string
}

// send message function to server
func (c *Client) SendMessage(messages chan string) {
	for{
		messages <- fmt.Sprintf("Hellow from client %d", c.Id) 
		time.Sleep(5 * time.Second) // every message is sent after every client Id seconds
	}
}

func (c *Client) ReceiveMessage(){
	for{
		msg := <- c.SubChannel
		fmt.Println(fmt.Sprintf("Message from Server at client %d: %s", c.Id, msg))
	}
}