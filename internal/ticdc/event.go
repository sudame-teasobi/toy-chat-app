package ticdc

import "encoding/json"

type OperationType string

const (
	OperationInsert OperationType = "INSERT"
	OperationUpdate OperationType = "UPDATE"
	OperationDelete OperationType = "DELETE"
)

type EventData struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Event struct {
	Version       int           `json:"version"`
	Database      string        `json:"database"`
	Table         string        `json:"table"`
	TableID       int           `json:"tableID"`
	Type          OperationType `json:"type"`
	CommitTs      int64         `json:"commitTs"`
	BuildTs       int64         `json:"buildTs"`
	SchemaVersion int64         `json:"schemaVersion"`
	Data          EventData     `json:"data"`
}
