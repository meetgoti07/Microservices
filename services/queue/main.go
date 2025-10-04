package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gin-quickstart/config"
	"gin-quickstart/database"
	"gin-quickstart/grpc"
	"gin-quickstart/kafka"
	"gin-quickstart/routes"
	"gin-quickstart/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Redis
	if err := database.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer database.CloseRedis()

	// Initialize gRPC Menu Service client
	menuClient, err := grpc.NewMenuClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Menu Service client: %v", err)
	} else {
		defer menuClient.Close()
		log.Println("Menu Service gRPC client initialized")
	}

	// Initialize Kafka Producer
	kafkaProducer, err := kafka.NewKafkaProducer(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
	} else {
		defer kafkaProducer.Close()
		log.Println("Kafka producer initialized")
	}

	// Initialize Queue Service
	queueService := services.NewQueueService()

	// Initialize and start Kafka Consumer
	kafkaConsumer, err := kafka.NewKafkaConsumer(cfg, queueService)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kafka consumer: %v", err)
	} else {
		if err := kafkaConsumer.Start(); err != nil {
			log.Printf("Warning: Failed to start Kafka consumer: %v", err)
		} else {
			defer kafkaConsumer.Stop()
			log.Println("Kafka consumer started successfully")
		}
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router)

	// Graceful shutdown
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		port := cfg.Port
		log.Printf("🚀 Queue service starting on port %s", port)
		log.Println("📊 Features enabled:")
		log.Println("  ✓ MySQL persistence")
		log.Println("  ✓ Redis real-time cache")
		log.Println("  ✓ Kafka event streaming")
		log.Println("  ✓ gRPC Menu Service client")
		log.Println("  ✓ Token-based queue system")
		log.Println("  ✓ Real-time position tracking")
		
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigint
	log.Println("🛑 Shutting down server...")

	// Cleanup
	if kafkaConsumer != nil {
		kafkaConsumer.Stop()
	}
	if kafkaProducer != nil {
		kafkaProducer.Close()
	}
	if menuClient != nil {
		menuClient.Close()
	}
	database.CloseRedis()
	database.Close()

	log.Println("✅ Server stopped gracefully")
	os.Exit(0)
}