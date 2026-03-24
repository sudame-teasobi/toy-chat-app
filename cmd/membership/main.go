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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/consumer"
	"github.com/sudame/chat/internal/handler"
	"github.com/sudame/chat/internal/infrastructure/query"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/internal/ticdc"
	"github.com/sudame/chat/pkg/env"
	"github.com/sudame/chat/pkg/httpclient"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	ctx := context.Background()

	serverPort := env.GetEnv("SERVER_PORT").Value()

	kafkaBroker := env.GetEnv("KAFKA_BROKER").WithDefault("localhost:9092").Value()
	kafkaGroupID := env.GetEnv("KAFKA_GROUP_ID").Value()

	host := env.GetEnv("DB_HOST").WithDefault("localhost").Value()
	port := env.GetEnv("DB_PORT").WithDefault("4000").Value()
	user := env.GetEnv("DB_USER").WithDefault("root").Value()
	password := env.GetEnv("DB_PASSWORD").WithDefault("").Value()
	dbName := env.GetEnv("DB_NAME").WithDefault("toy_chat_app").Value()

	roomServerBaseURL := env.GetEnv("ROOM_SERVER").Value()

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
	httpClient := httpclient.NewHTTPClient(client, roomServerBaseURL)

	roomQuery := query.NewRoomQuery(httpClient)

	createMembershipService := service.NewCreateMembershipService(userRepo, roomQuery, membershipRepo)

	membershipConsumer := consumer.NewMembershipConsumer(createMembershipService)

	checkMembershipExistenceService := service.NewCheckMembershipExistenceService(membershipRepo)
	checkMembershipExistenceHandler := handler.NewCheckMembershipExistenceHandler(checkMembershipExistenceService)

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Printf("%s %s %d %v - error: %v", v.Method, v.URI, v.Status, v.Latency, v.Error)
			} else {
				log.Printf("%s %s %d %v", v.Method, v.URI, v.Status, v.Latency)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.POST(query.CheckMembershipExistencePath, checkMembershipExistenceHandler.Handle)

	go func() {
		log.Printf("Starting server on port %s", serverPort)
		if err := e.Start(":" + serverPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{kafkaBroker},
			GroupID:     kafkaGroupID,
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
	slog.DebugContext(ctx, "processing message",
		slog.String("topic", m.Topic),
		slog.Int("partition", m.Partition),
		slog.Int64("offset", m.Offset),
		slog.String("key", string(m.Key)),
		slog.String("value", string(m.Value)),
	)
	var e ticdc.Event
	if err := json.Unmarshal(m.Value, &e); err != nil {
		return fmt.Errorf("failed to unmarshal data on kafka: %w", err)
	}
	if err := membershipConsumer.Consume(ctx, e); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}
