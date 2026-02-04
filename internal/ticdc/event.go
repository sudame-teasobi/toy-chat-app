package ticdc

import "encoding/json"

type OperationType string

const (
	OperationInsert OperationType = "INSERT"
	OperationUpdate OperationType = "UPDATE"
	OperationDelete OperationType = "DELETE"
)

type EventData struct {
	ID      string          `json:"id"`
	Type    string          `json:"event_type"`
	Payload json.RawMessage `json:"payload"`
}

type Event struct {
	Data []EventData `json:"data"`
}
