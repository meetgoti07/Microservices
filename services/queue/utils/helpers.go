package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gin-quickstart/database"
	"gin-quickstart/models"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateTokenNumber generates a sequential token number
func GenerateTokenNumber(db interface{}) (string, error) {
	// Implementation for token generation
	today := time.Now().UTC().Truncate(24 * time.Hour)
	
	var counter models.QueueTokenCounter
	result := database.GetDB().Where("date = ?", today).First(&counter)
	
	if result.Error != nil {
		// Create new counter for today
		counter = models.QueueTokenCounter{
			ID:            GenerateUUID(),
			Date:          today,
			CurrentNumber: 1,
			Prefix:        "A",
			LastResetAt:   time.Now().UTC(),
		}
		database.GetDB().Create(&counter)
		return fmt.Sprintf("%s%03d", counter.Prefix, counter.CurrentNumber), nil
	}
	
	// Increment counter
	counter.CurrentNumber++
	database.GetDB().Save(&counter)
	
	return fmt.Sprintf("%s%03d", counter.Prefix, counter.CurrentNumber), nil
}

// CacheQueueEntry caches queue entry in Redis
func CacheQueueEntry(ctx context.Context, entry *models.QueueEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	key := fmt.Sprintf("queue:entry:%s", entry.ID)
	return database.GetRedis().Set(ctx, key, data, 1*time.Hour).Err()
}

// GetCachedQueueEntry retrieves cached queue entry from Redis
func GetCachedQueueEntry(ctx context.Context, entryID string) (*models.QueueEntry, error) {
	key := fmt.Sprintf("queue:entry:%s", entryID)
	data, err := database.GetRedis().Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	var entry models.QueueEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, err
	}
	
	return &entry, nil
}

// InvalidateQueueCache invalidates queue cache
func InvalidateQueueCache(ctx context.Context, entryID string) error {
	key := fmt.Sprintf("queue:entry:%s", entryID)
	return database.GetRedis().Del(ctx, key).Err()
}

// CalculateEstimatedWaitTime calculates estimated wait time based on position
func CalculateEstimatedWaitTime(position int, avgPrepTimePerItem int, bufferTime int) int {
	return (position * avgPrepTimePerItem) + bufferTime
}

// CalculateEstimatedReadyTime calculates estimated ready time
func CalculateEstimatedReadyTime(estimatedWaitTime int) time.Time {
	return time.Now().UTC().Add(time.Duration(estimatedWaitTime) * time.Minute)
}

// StringPtr returns pointer to string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns pointer to int
func IntPtr(i int) *int {
	return &i
}

// TimePtr returns pointer to time
func TimePtr(t time.Time) *time.Time {
	return &t
}
