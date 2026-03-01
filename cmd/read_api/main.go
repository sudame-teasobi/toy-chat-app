package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sudame/chat/internal/read_api/graph"
	"github.com/sudame/chat/internal/read_api/middleware"
	"github.com/sudame/chat/internal/read_api/resolver"
	"github.com/sudame/chat/pkg/env"
)

func main() {
	ctx := context.Background()

	port := env.GetEnv("SERVER_PORT").Value()
	dynamodbURL := env.GetEnv("DYNAMODB_URL").Value()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ap-northeast-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
	)
	if err != nil {
		log.Fatalf("failed to load default config: %s", err)
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = &dynamodbURL
	})

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &resolver.Resolver{
			DynamoDBClient: dynamodbClient,
		},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	http.Handle("/query", middleware.AuthMiddleware(srv))

	log.Printf("GraphQL playground: http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
