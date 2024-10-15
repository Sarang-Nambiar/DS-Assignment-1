package client

type Message struct{
	Clock []int
	Message string
	ClientId int
}

func (m Message) IsEmpty() bool {
	return m.Clock == nil && m.Message == "" && m.ClientId == 0
}