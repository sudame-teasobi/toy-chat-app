package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service/membership"
)

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx := context.Background()

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

	membershipRepository := repository.NewMembershipRepository(db)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{"localhost:9092"},
			GroupID:     "membership-consumer-group",
			GroupTopics: []string{"chat-room-events"},
		},
	)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Error: %v", err)
		}

		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s/n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		err = membership.Listen(ctx, membershipRepository, m.Value)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
	}
}
