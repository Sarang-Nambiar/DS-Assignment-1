package server

import (
	"fmt"
	"math/rand"
	"time"
	"strconv"
)

type Server struct {
	Drop bool
	ClientChannels []chan string
}

// function to receive messages from client
func (s *Server) ReceiveMessage(messages chan string){
	for{
		msg := <- messages
		fmt.Println(fmt.Sprintf("Received message at server: %s", msg))

		if msg != "" {
			clientId, err := strconv.ParseUint(string(msg[len(msg)-1]), 10, 8)
			if err != nil {
				fmt.Println("Error parsing client ID:", err)
				continue
			}
			// send to all clients which aren't the clientId
			s.sendMessage(int(clientId), msg)
		}
	}
}

// function to send message to clients except the one who sent the message
func (s *Server) sendMessage(clientId int, message string){
	
	if !s.coinFlip(){
		return
	}

	for i := 0; i < 10; i++ {
		if i != clientId{
			s.ClientChannels[i] <- message
			fmt.Println(fmt.Sprintf("Message sent to client %d: %s", i, message))
		}
	}
}

// coin flip to decide if the server should drop the message
func (s *Server) coinFlip() bool{
	rand.Seed(time.Now().UnixNano()) // Making sure this is random using a unique seed
	return rand.Intn(2) == 1 // Generates random number from 0 to 1
}