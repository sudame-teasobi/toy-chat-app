package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/consumer"
	"github.com/sudame/chat/internal/infrastructure/query"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
	"github.com/sudame/chat/pkg/httpclient"
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

	roomServerBaseURL := getEnv("ROOM_SERVER", "")

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
	membershipRepo := repository.NewMembershipRepository(db)

	client := &http.Client{}
	httpClient := httpclient.NewHTTPClient(client)

	roomQuery := query.NewRoomQuery(httpClient, roomServerBaseURL)

	createMembershipService := service.NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

	membershipConsumer := consumer.NewMembershipConsumer(createMembershipService)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{kafkaBroker},
			GroupID:     "membership-consumer-group",
			GroupTopics: []string{"event-records-changefeed"},
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

		m, err := reader.FetchMessage(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch message", "err", err)
			continue
		}

		if err := process(ctx, m, membershipConsumer); err != nil {
			slog.ErrorContext(ctx, "failed to handle event", "err", err)
			continue
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			slog.ErrorContext(ctx, "failed to commit message", "err", err)
		}
	}

}

func process(ctx context.Context, m kafka.Message, membershipConsumer *consumer.MembershipConsumer) error {
	log.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	var e ticdc.Event
	if err := json.Unmarshal(m.Value, &e); err != nil {
		return fmt.Errorf("failed to unmarshal data on kafka: %w", err)
	}
	if err := membershipConsumer.Consume(ctx, e); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}
