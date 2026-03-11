package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sudame/chat/internal/handler"
	"github.com/sudame/chat/internal/infrastructure/query"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/pkg/env"
	"github.com/sudame/chat/pkg/httpclient"
)

func main() {
	ctx := context.Background()

	serverPort := env.GetEnv("SERVER_PORT").Value()

	host := env.GetEnv("DB_HOST").WithDefault("localhost").Value()
	port := env.GetEnv("DB_PORT").WithDefault("4000").Value()
	user := env.GetEnv("DB_USER").WithDefault("root").Value()
	password := env.GetEnv("DB_PASSWORD").WithDefault("").Value()
	dbName := env.GetEnv("DB_NAME").WithDefault("toy_chat_app").Value()

	membershipServerBaseURL := env.GetEnv("MEMBERSHIP_SERVER").Value()

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

	messageRepo := repository.NewMessageRepository(db)

	client := &http.Client{}
	httpClient := httpclient.NewHTTPClient(client, membershipServerBaseURL)

	membershipQuery := query.NewMembershipQuery(httpClient)

	service := service.NewPostMessageService(ctx, membershipQuery, messageRepo)
	handler := handler.NewPostMessageHandler(service)

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

	e.POST("/post-message", handler.Handle)

	log.Printf("Starting server on port %s", serverPort)
	if err := e.Start(":" + serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
