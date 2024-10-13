package server

import (
	"fmt"
)

type Server struct {
	drop bool
	clientList [10]uint8
}

// function to receive messages from client
func (s *Server) receiveMessage(messages chan string){
	fmt.Println("Receiving message from client")
}

// function to send message to clients except the one who sent the message
func (s *Server) sendMessage(){
	fmt.Println("Sending message to clients except the one who sent the message")
}

// coin flip to decide if the server should drop the message
func (s *Server) coinFlip(){
	fmt.Println("Coin flip")
}