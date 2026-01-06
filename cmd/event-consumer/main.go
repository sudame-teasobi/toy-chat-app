package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/applicationservice"
	"github.com/sudame/chat/internal/domain/chatroom"
	"github.com/sudame/chat/internal/infrastructure/repository"
)

// CanalJSONMessage represents a TiCDC canal-json format message.
type CanalJSONMessage struct {
	ID       int64             `json:"id"`
	Database string            `json:"database"`
	Table    string            `json:"table"`
	PKNames  []string          `json:"pkNames"`
	IsDDL    bool              `json:"isDdl"`
	Type     string            `json:"type"`
	Data     []EventRecordData `json:"data"`
}

// EventRecordData represents a row from event_records table.
type EventRecordData struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

// ChatRoomCreatedPayload is the payload of ChatRoomCreatedEvent.
type ChatRoomCreatedPayload struct {
	ChatRoomID    string `json:"ChatRoomID"`
	Name          string `json:"Name"`
	CreatorUserID string `json:"CreatorUserID"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutdown signal received")
		cancel()
	}()

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "4000")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "toy_chat_app")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	chatRoomRepo := repository.NewChatRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize use case
	addMemberUsecase := applicationservice.NewAddChatRoomMemberUsecase(chatRoomRepo, userRepo)

	// Kafka consumer configuration
	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	topic := getEnv("KAFKA_TOPIC", "event-records-changefeed")
	groupID := getEnv("KAFKA_GROUP_ID", "event-consumer-group")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Printf("Starting event consumer (broker=%s, topic=%s, group=%s)", kafkaBroker, topic, groupID)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer")
			return
		default:
		}

		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		if err := processMessage(ctx, msg.Value, addMemberUsecase); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}
}

func processMessage(ctx context.Context, value []byte, addMemberUsecase *applicationservice.AddChatRoomMemberUsecase) error {
	var canalMsg CanalJSONMessage
	if err := json.Unmarshal(value, &canalMsg); err != nil {
		return fmt.Errorf("failed to unmarshal canal message: %w", err)
	}

	// Only process INSERT events
	if canalMsg.Type != "INSERT" {
		return nil
	}

	// Only process event_records table
	if canalMsg.Table != "event_records" {
		return nil
	}

	for _, record := range canalMsg.Data {
		if err := handleEventRecord(ctx, record, addMemberUsecase); err != nil {
			log.Printf("Error handling event record %s: %v", record.ID, err)
		}
	}

	return nil
}

func handleEventRecord(ctx context.Context, record EventRecordData, addMemberUsecase *applicationservice.AddChatRoomMemberUsecase) error {
	switch record.EventType {
	case "ChatRoomCreated":
		return handleChatRoomCreated(ctx, record.Payload, addMemberUsecase)
	default:
		// Ignore other event types for now
		return nil
	}
}

func handleChatRoomCreated(ctx context.Context, payloadStr string, addMemberUsecase *applicationservice.AddChatRoomMemberUsecase) error {
	var payload ChatRoomCreatedPayload
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal ChatRoomCreated payload: %w", err)
	}

	// CreatorUserIDが空の場合はスキップ（旧形式のイベント）
	if payload.CreatorUserID == "" {
		log.Printf("Skipping ChatRoomCreated event without CreatorUserID: room=%s", payload.ChatRoomID)
		return nil
	}

	log.Printf("Processing ChatRoomCreated event: room=%s, creator=%s", payload.ChatRoomID, payload.CreatorUserID)

	// Add creator as member
	_, err := addMemberUsecase.Execute(ctx, applicationservice.AddChatRoomMemberInput{
		ChatRoomID: payload.ChatRoomID,
		UserID:     payload.CreatorUserID,
	})
	if err != nil {
		// 既にメンバーの場合は無視（移行期間中の並行稼働対応）
		if errors.Is(err, chatroom.ErrAlreadyMember) {
			log.Printf("Creator %s is already a member of room %s (skipping)", payload.CreatorUserID, payload.ChatRoomID)
			return nil
		}
		return fmt.Errorf("failed to add creator as member: %w", err)
	}

	log.Printf("Successfully added creator %s as member of room %s", payload.CreatorUserID, payload.ChatRoomID)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}