package client

import (
	"fmt"
	"time"
)

type Client struct{
	Id int
	SendChannel chan Message
	ReceiveChannel chan Message
	Clock int
}

// send message function to server
func (c *Client) SendMessage() {
	for{
		c.Clock += 1
		message := Message{c.Clock, fmt.Sprintf("Hello from client %d", c.Id), c.Id}
		fmt.Println(fmt.Sprintf("[CLIENT-%d-L%d] Sending message to server: '%s'", c.Id, c.Clock, message.Message))
		c.SendChannel <- message
		time.Sleep(5 * time.Second) // each message is sent every 5 seconds
	}
}

func (c *Client) ReceiveMessage(){
	for{
		msg := <- c.ReceiveChannel
		c.Clock = max(c.Clock, msg.Clock) + 1 // updating the logical clock by finding the maximum between the two clock values
		fmt.Println(fmt.Sprintf("[CLIENT-%d-L%d] Message received from server: '%s'", c.Id, c.Clock, msg.Message))
	}
}
