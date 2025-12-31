package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
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
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to TiDB: %v", err)
	}

	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		log.Fatalf("Failed to query version: %v", err)
	}

	fmt.Printf("Connected to TiDB successfully!\n")
	fmt.Printf("TiDB version: %s\n", version)

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_write (
			id INT AUTO_INCREMENT PRIMARY KEY,
			message VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create test table: %v", err)
	}
	fmt.Println("Test table created/verified.")

	// Insert test data
	result, err := db.Exec("INSERT INTO test_write (message) VALUES (?)", "Hello from write-server!")
	if err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}
	insertedID, _ := result.LastInsertId()
	fmt.Printf("Inserted test data with ID: %d\n", insertedID)

	// Verify the inserted data
	var message string
	err = db.QueryRow("SELECT message FROM test_write WHERE id = ?", insertedID).Scan(&message)
	if err != nil {
		log.Fatalf("Failed to read inserted data: %v", err)
	}
	fmt.Printf("Verified inserted data: %q\n", message)

	// Delete the test data
	result, err = db.Exec("DELETE FROM test_write WHERE id = ?", insertedID)
	if err != nil {
		log.Fatalf("Failed to delete test data: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Deleted %d row(s).\n", rowsAffected)

	// Verify deletion
	err = db.QueryRow("SELECT message FROM test_write WHERE id = ?", insertedID).Scan(&message)
	if err == sql.ErrNoRows {
		fmt.Println("Verified: data was successfully deleted.")
	} else if err != nil {
		log.Fatalf("Failed to verify deletion: %v", err)
	} else {
		log.Fatalf("Data still exists after deletion!")
	}

	fmt.Println("All write/delete tests passed!")

	// ==========================================
	// Kafka connectivity test
	// ==========================================
	fmt.Println("\n--- Kafka Connectivity Test ---")

	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	kafkaTopic := "test-write-server"

	// Connect to Kafka and get broker metadata
	conn, err := kafka.Dial("tcp", kafkaBroker)
	if err != nil {
		log.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer conn.Close()

	// Get broker info
	brokers, err := conn.Brokers()
	if err != nil {
		log.Fatalf("Failed to get Kafka brokers: %v", err)
	}
	fmt.Printf("Connected to Kafka successfully!\n")
	fmt.Printf("Brokers: ")
	for i, b := range brokers {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s:%d", b.Host, b.Port)
	}
	fmt.Println()

	// Create test topic (if not exists)
	controller, err := conn.Controller()
	if err != nil {
		log.Fatalf("Failed to get Kafka controller: %v", err)
	}
	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		log.Fatalf("Failed to connect to Kafka controller: %v", err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             kafkaTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		log.Printf("Topic creation (may already exist): %v", err)
	}
	fmt.Printf("Test topic '%s' ready.\n", kafkaTopic)

	// Produce a test message
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaBroker),
		Topic:        kafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
	}
	defer writer.Close()

	testMessage := fmt.Sprintf("Hello from write-server at %s", time.Now().Format(time.RFC3339))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte(testMessage),
	})
	if err != nil {
		log.Fatalf("Failed to produce message to Kafka: %v", err)
	}
	fmt.Printf("Produced message: %q\n", testMessage)

	// Consume the test message
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker},
		Topic:     kafkaTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
		MaxWait:   3 * time.Second,
	})
	defer reader.Close()

	// Seek to the latest offset minus 1 to read our message
	reader.SetOffset(kafka.LastOffset)
	// Read from beginning to find our message
	reader.SetOffset(0)

	readCtx, readCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer readCancel()

	var foundMessage bool
	for {
		msg, err := reader.ReadMessage(readCtx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			log.Fatalf("Failed to consume message from Kafka: %v", err)
		}
		if string(msg.Value) == testMessage {
			fmt.Printf("Consumed message: %q\n", string(msg.Value))
			foundMessage = true
			break
		}
	}

	if !foundMessage {
		log.Fatalf("Failed to find the produced message in Kafka")
	}

	fmt.Println("All Kafka tests passed!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
