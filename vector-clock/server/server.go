package server

import (
	"fmt"
	"vector-clock/client"
	"math/rand"
	"sync"
	"time"
)


type Server struct {
	Clock []int
	SendChannels []chan client.Message
	ReceiveChannels []chan client.Message
	Lock sync.Mutex
}

// function to receive messages from client
func (s *Server) ReceiveMessage(){
	for i := range s.ReceiveChannels{
		go s.handleClientChannels(i)
	}
}

// function to handle all client channels
func (s *Server) handleClientChannels(clientId int){
	for{
		msg := <- s.ReceiveChannels[clientId]
		
		if client.CausalityDetection(msg.Clock, s.Clock) {
			fmt.Println(fmt.Sprintf("[SERVER-VC%v] Potential Causality Violation detected for message: '%s'. Message Clock: %v", s.Clock, msg.Message, msg.Clock))
		}

		s.Lock.Lock()
		s.Clock = client.VectorMAX(s.Clock, msg.Clock) // updating the logical clock by finding the maximum between the two clock values
		s.Clock[len(s.Clock) - 1] += 1
		fmt.Println(fmt.Sprintf("[SERVER-VC%v] Message receieved: '%s'", s.Clock, msg.Message))
		s.Lock.Unlock()

		if !msg.IsEmpty() {
			// send to all clients which don't have id as clientId
			s.sendMessage(msg)
		}
	}
}

// function to send message to clients except the one who sent the message
func (s *Server) sendMessage(message client.Message){
	
	if !s.coinFlip(){
		s.Lock.Lock()
		s.Clock[len(s.Clock) - 1] += 1
		currentClock := s.Clock
		s.Lock.Unlock()

		fmt.Println(fmt.Sprintf("[SERVER-VC%v] Forwarding the message of client %d is dropped", currentClock, message.ClientId))
		return
	}

	for i, channel := range s.SendChannels{
		if i != message.ClientId{

			s.Lock.Lock()
			s.Clock[len(s.Clock) - 1] += 1
			currentClock := s.Clock
			s.Lock.Unlock()
			
			channel <- client.Message{currentClock, message.Message, message.ClientId}
			fmt.Println(fmt.Sprintf("[SERVER-VC%v] Message forwarded to client %d: '%s'", s.Clock, i, message.Message))
		}
	}
}

// coin flip to decide if the server should drop the message
func (s *Server) coinFlip() bool{
	rand.Seed(time.Now().UnixNano()) // Making sure this is random using a unique seed
	return rand.Intn(2) == 1 // Generates random number from 0 to 1
}
