package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sudame/chat/internal/handler"
	"github.com/sudame/chat/internal/infrastructure/repository"
	"github.com/sudame/chat/internal/service"
	"github.com/sudame/chat/pkg/env"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	host := env.GetEnv("DB_HOST").WithDefault("localhost").Value()
	port := env.GetEnv("DB_PORT").WithDefault("4000").Value()
	user := env.GetEnv("DB_USER").WithDefault("root").Value()
	password := env.GetEnv("DB_PASSWORD").WithDefault("").Value()
	dbName := env.GetEnv("DB_NAME").WithDefault("toy_chat_app").Value()

	serverPort := env.GetEnv("SERVER_PORT").WithDefault("8080").Value()

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

	roomRepo := repository.NewChatRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

	createUserService := service.NewCreateUserService(userRepo)
	createRoomService := service.NewCreateRoomService(userRepo, roomRepo)
	checkRoomExistenceService := service.NewCheckRoomExistenceService(roomRepo)

	createUserHandler := handler.NewCreateUserHandler(createUserService)
	createRoomHandler := handler.NewCreateRoomHandler(createRoomService)
	checkRoomExistenceHandler := handler.NewCheckRoomExistenceHandler(checkRoomExistenceService)

	e.POST("/create-user", createUserHandler.Handle)
	e.POST("/create-chat-room", createRoomHandler.Handle)
	e.POST("/check-room-existence", checkRoomExistenceHandler.Handle)

	log.Printf("Starting server on port %s", serverPort)
	if err := e.Start(":" + serverPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
