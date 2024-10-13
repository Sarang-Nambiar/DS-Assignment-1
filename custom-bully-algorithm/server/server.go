package server

import (
	"fmt"
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
		fmt.Println(fmt.Sprintf("Received message: %s", msg))

		if msg != "" {
			clientId, err := strconv.ParseUint(string(msg[len(msg)-1]), 10, 8)
			if err != nil {
				fmt.Println("Error parsing client ID:", err)
				continue
			}
			// send to all clients which aren't the clientId
			s.SendMessage(int(clientId), msg)
		}
	}
}

// function to send message to clients except the one who sent the message
func (s *Server) SendMessage(clientId int, message string){
	for i := 0; i < 10; i++ {
		if i != clientId{
			s.ClientChannels[clientId] <- message
			fmt.Println(fmt.Sprintf("Message sent to client %d: %s", clientId, message))
		}
	}
}

// coin flip to decide if the server should drop the message
func (s *Server) CoinFlip(){
	fmt.Println("Coin flip")
}