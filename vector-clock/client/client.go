package client

import (
	"fmt"
	"time"
)

type Client struct{
	Id int
	SendChannel chan Message
	ReceiveChannel chan Message
	Clock []int
}

// send message function to server
func (c *Client) SendMessage() {
	for{
		c.Clock[c.Id] += 1
		message := Message{c.Clock, fmt.Sprintf("Hello from client %d", c.Id), c.Id}
		fmt.Println(fmt.Sprintf("[CLIENT-%d-VC%v] Sending message to server: '%s'", c.Id, c.Clock, message.Message))
		c.SendChannel <- message
		time.Sleep(5 * time.Second) // each message is sent every 5 seconds
	}
}

func (c *Client) ReceiveMessage(){
	for{
		msg := <- c.ReceiveChannel

		// Causality Detections
		if CausalityDetection(msg.Clock, c.Clock) {
			fmt.Println(fmt.Sprintf("[CLIENT-%d-VC%v]Potential Causality Violation detected for message: '%s'. Message Clock: %v", c.Id, c.Clock, msg.Message, msg.Clock))
		}

		// updating the logical clock by finding the maximum between the two clock values
		c.Clock = VectorMAX(c.Clock, msg.Clock)
		c.Clock[c.Id] += 1
		fmt.Println(fmt.Sprintf("[CLIENT-%d-VC%v] Message received from server: '%s'", c.Id, c.Clock, msg.Message))
	}
}

// Utility functions for vector clocks

func VectorMAX(clock1 []int, clock2 []int) []int {
	for i := range clock1{
		clock1[i] = max(clock1[i], clock2[i])
	}
	return clock1
}

// Checking for causality violations
func CausalityDetection(msgClock []int, localClock []int) bool {
	for i := range msgClock{
		if msgClock[i] > localClock[i]{
			return false
		}
	}
	return true
}
