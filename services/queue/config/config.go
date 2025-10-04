package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Kafka
	KafkaBrokers []string
	KafkaGroupID string

	// Auth Service
	AuthServiceURL string

	// gRPC Menu Service
	MenuServiceHost string
	MenuServicePort string

	// Queue Configuration
	MaxConcurrentOrders          int
	AvgPreparationTimePerItem    int
	BufferTime                   int
	ExpressQueueMaxItems         int
	MaxWaitTimeAlert             int
	TokenExpiryTime              int
	NotificationPositionThreshold int
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "3004"),

		DBHost:     getEnv("DB_HOST", "mysql"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "root"),
		DBName:     getEnv("DB_NAME", "queue_db"),

		RedisHost:     getEnv("REDIS_HOST", "redis"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "kafka:9092")},
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "queue-service-group"),

		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://auth-service:3001"),

		MenuServiceHost: getEnv("MENU_SERVICE_HOST", "menu-service"),
		MenuServicePort: getEnv("MENU_SERVICE_PORT", "50051"),

		MaxConcurrentOrders:          getEnvAsInt("MAX_CONCURRENT_ORDERS", 10),
		AvgPreparationTimePerItem:    getEnvAsInt("AVG_PREP_TIME_PER_ITEM", 5),
		BufferTime:                   getEnvAsInt("BUFFER_TIME", 2),
		ExpressQueueMaxItems:         getEnvAsInt("EXPRESS_QUEUE_MAX_ITEMS", 3),
		MaxWaitTimeAlert:             getEnvAsInt("MAX_WAIT_TIME_ALERT", 30),
		TokenExpiryTime:              getEnvAsInt("TOKEN_EXPIRY_TIME", 60),
		NotificationPositionThreshold: getEnvAsInt("NOTIFICATION_POSITION_THRESHOLD", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
