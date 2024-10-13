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
		time.Sleep(10 * time.Second) // every two seconds, a message is sent
	}
}

func (c *Client) ReceiveMessage(){
	for{
		msg := <- c.SubChannel
		fmt.Println(fmt.Sprintf("Message from Server: %s", msg))
	}
}