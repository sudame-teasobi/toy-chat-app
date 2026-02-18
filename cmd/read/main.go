package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/segmentio/kafka-go"
	"github.com/sudame/chat/internal/consumer"
	"github.com/sudame/chat/internal/ticdc"
	"github.com/sudame/chat/pkg/env"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	dynamodbURL, err := env.GetEnv("DYNAMODB_URL").Value()
	if err != nil {
		panic("failed to get env: DYNAMODB_URL")
	}

	kafkaBroker, err := env.GetEnv("KAFKA_BROKER").WithDefault("localhost:9092").Value()
	if err != nil {
		log.Fatalf("failed to get env KAFKA_BROKER: %s", err)
	}

	slog.DebugContext(ctx, "dynamodb", "dynamodb_url", dynamodbURL)

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ap-northeast-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	if err != nil {
		log.Fatalf("failed to load default config: %s", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = &dynamodbURL
	})

	err = initializeTables(ctx, client)
	if err != nil {
		log.Fatalf("failed to initialize tables: %s", err)
	}

	readModelConsumer := consumer.NewReadModelConsumer(client)

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{kafkaBroker},
			GroupID:     "read-model-consumer-group",
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
		err = readModelConsumer.Consume(ctx, e)
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}

}
