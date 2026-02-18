package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

	tables, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("failed to list tables: %s", err)
	}

	slog.DebugContext(ctx, "tables", "tables", tables)
}
