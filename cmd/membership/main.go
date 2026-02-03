package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/consumer"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
)

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx := context.Background()

	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "4000")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "toy_chat_app")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to TiDB: %v", err)
	}
	log.Println("Connected to TiDB successfully")

	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewChatRoomRepository(db)
	membershipRepo := repository.NewMembershipRepository(db)

	createMembershipService := service.NewCreateMembershipService(userRepo, roomRepo, membershipRepo)

	membershipConsumer := consumer.NewMembershipConsumer(createMembershipService)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{kafkaBroker},
			GroupID:     "membership-consumer-group",
			GroupTopics: []string{"chat-room-events"},
		},
	)

	defer func() {
		err := reader.Close()
		if err != nil {
			log.Printf("failed to close reader: %s", err.Error())
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		default:
		}

		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			continue
		}

		log.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		var e ticdc.Event
		err = json.Unmarshal(m.Value, &e)
		if err != nil {
			log.Printf("failed to unmarshal data on kafka: %s", err.Error())
			continue
		}
		err = membershipConsumer.Consume(ctx, e)
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}
}
