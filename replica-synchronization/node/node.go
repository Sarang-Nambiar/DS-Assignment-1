package node

import (
	"fmt"
	"sync"
	"time"
)

type Node struct{
	Id int
	LocalReplica string
	Ring []int
	IsCoordinator bool
	CoordinatorId int
	Channels []chan Message
	Lock sync.Mutex
}

// TODO: Need to check when the last time the local replica was synchronized to know if the coordinator is dead or not

func (n *Node) SendMessage(){
	if n.IsCoordinator {
		for {
			fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Sync has begun.", n.CoordinatorId))
			for i := range n.Channels {
				n.Channels[i] <- Message{"PUT", n.CoordinatorId, n.LocalReplica}
				fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Replica: '%s', sent to node %d", n.CoordinatorId, n.LocalReplica, i))
			}

			time.Sleep(5 * time.Second) // Synchronizes after 5 seconds
		}
	} else {
		n.Channels[n.Id] <- Message{"ACK", n.Id, ""}
		fmt.Println(fmt.Sprintf("[NODE-%d] ACK message sent to Coordinator", n.Id))
	}
}

func (n *Node) ReceiveMessage(){
	if n.IsCoordinator {
		for i := range n.Channels{
			go n.handleNodeChannels(i)
		}
	} else{
		go n.handleNodeChannels(n.Id)
	}
}

func (n *Node) handleNodeChannels(nodeId int) {
	for{
		select {
		case msg := <- n.Channels[nodeId]:
			if !msg.IsEmpty() {
				switch(msg.Type) {
				case "ACK":
					// Check to see which nodes received the message and which didn't
					fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Replica synchornized with node %d", n.CoordinatorId, nodeId))			
				case "PUT":
					select {
					case syncMsg := <- n.Channels[nodeId]:
						fmt.Println(fmt.Sprintf("[NODE-%d] Replica: '%s', received from Coordinator %d", n.Id, syncMsg.Payload, syncMsg.nodeId))
						
						n.Lock.Lock()
						n.LocalReplica = syncMsg.Payload
						n.Lock.Unlock()
		
						// Sending acknowledgement message to coordinator
						n.SendMessage()
					case <- time.After(6 * time.Second):
						fmt.Println(fmt.Sprintf("[NODE-%d] No message from the Coordinator %d, initiating coordinator ring election", n.Id, n.CoordinatorId))
					}
				default:
					fmt.Println("Error 404. Unknown Message Type")
				}
			}
		case <- time.After(6 * time.Second):
			fmt.Println(fmt.Sprintf("[COORDINATOR-%d] Node %d didn't acknowledge the sync message", n.CoordinatorId, nodeId))
		}
	}
}