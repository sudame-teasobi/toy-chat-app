package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	baseURL      = "http://localhost:8300/api/v2/changefeeds"
	changefeedID = "event-records-changefeed"
)

type ChangefeedConfig struct {
	ChangefeedID  string        `json:"changefeed_id"`
	ReplicaConfig ReplicaConfig `json:"replica_config"`
	SinkURI       string        `json:"sink_uri"`
}

type ReplicaConfig struct {
	Filter Filter `json:"filter"`
}

type Filter struct {
	Rules []string `json:"rules"`
}

func main() {
	cmd := flag.String("cmd", "", "command to execute: list, create, or delete (required)")
	flag.Parse()

	if *cmd == "" {
		log.Fatal("-cmd is required (use 'list', 'create', or 'delete')")
	}

	var req *http.Request
	var err error

	switch *cmd {
	case "list":
		req, err = listChangefeedsRequest()
	case "create":
		req, err = createChangefeedRequest()
	case "delete":
		req, err = deleteChangefeedRequest()
	default:
		log.Fatalf("unknown command: %s (use 'list', 'create', or 'delete')", *cmd)
	}

	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
	}

	fmt.Printf("Status: %s\n\n", res.Status)

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Body: %s\n", string(body))
	} else {
		fmt.Println(prettyJSON.String())
	}
}

func createChangefeedRequest() (*http.Request, error) {
	config := ChangefeedConfig{
		ChangefeedID: changefeedID,
		ReplicaConfig: ReplicaConfig{
			Filter: Filter{
				Rules: []string{"toy_chat_app.event_records"},
			},
		},
		SinkURI: "kafka://kafka:29092/event-records-changefeed?protocol=canal-json&enable-tidb-extension=false",
	}

	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func deleteChangefeedRequest() (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", baseURL, changefeedID)
	return http.NewRequest(http.MethodDelete, url, nil)
}

func listChangefeedsRequest() (*http.Request, error) {
	return http.NewRequest(http.MethodGet, baseURL, nil)
}
