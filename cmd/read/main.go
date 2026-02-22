package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	dynamodbURL, err := env.GetEnv("DYNAMODB_URL").SafeValue()
	if err != nil {
		panic("failed to get env: DYNAMODB_URL")
	}

	kafkaBroker, err := env.GetEnv("KAFKA_BROKER").WithDefault("localhost:9092").SafeValue()
	if err != nil {
		log.Fatalf("failed to get env KAFKA_BROKER: %s", err)
	}

	kafkaGroupID := env.GetEnv("KAFKA_GROUP_ID").Value()

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

		if err := process(ctx, m, readModelConsumer); err != nil {
			slog.ErrorContext(ctx, "failed to handle event", "err", err)
			continue
		}

		if err := reader.CommitMessages(ctx, m); err != nil {
			slog.ErrorContext(ctx, "failed to commit message", "err", err)
		}
	}

}

func process(ctx context.Context, m kafka.Message, readModelConsumer *consumer.ReadModelConsumer) error {
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
	if err := readModelConsumer.Consume(ctx, e); err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}

	return nil
}
