package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sudame/chat/internal/handler"
	"github.com/sudame/chat/internal/infrastructure/repository"
)

func main() {
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

	chatRoomRepo := repository.NewChatRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

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

	createChatRoomHandler := handler.NewCreateChatRoomHandler(chatRoomRepo, userRepo)
	e.POST("/create-chat-room", createChatRoomHandler.Handle)

	createUserHandler := handler.NewCreateUserHandler(userRepo)
	e.POST("/create-user", createUserHandler.Handle)

	serverPort := getEnv("SERVER_PORT", "8080")
	log.Printf("Starting server on port %s", serverPort)
	if err := e.Start(":" + serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
