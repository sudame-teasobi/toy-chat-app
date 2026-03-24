package model

type Message struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

func (Message) IsNode()         {}
func (m Message) GetID() string { return m.ID }
