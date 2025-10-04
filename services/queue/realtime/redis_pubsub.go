package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gin-quickstart/database"
	"gin-quickstart/models"

	"github.com/redis/go-redis/v9"
)

const (
	QueueUpdatesChannel = "queue:updates"
	QueueStatsChannel   = "queue:stats"
)

type RealtimeService struct {
	redis *redis.Client
}

func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		redis: database.GetRedis(),
	}
}

// PublishQueueUpdate publishes queue update to Redis pub/sub
func (rs *RealtimeService) PublishQueueUpdate(ctx context.Context, entry *models.QueueEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal queue entry: %w", err)
	}

	if err := rs.redis.Publish(ctx, QueueUpdatesChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish queue update: %w", err)
	}

	log.Printf("Published queue update: token=%s, position=%d, status=%s",
		entry.TokenNumber, entry.Position, entry.Status)

	return nil
}

// PublishQueueStats publishes queue statistics
func (rs *RealtimeService) PublishQueueStats(ctx context.Context, stats interface{}) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal queue stats: %w", err)
	}

	if err := rs.redis.Publish(ctx, QueueStatsChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish queue stats: %w", err)
	}

	log.Println("Published queue stats update")
	return nil
}

// SubscribeQueueUpdates subscribes to queue updates
func (rs *RealtimeService) SubscribeQueueUpdates(ctx context.Context, callback func(*models.QueueEntry)) error {
	pubsub := rs.redis.Subscribe(ctx, QueueUpdatesChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Println("Subscribed to queue updates channel")

	for {
		select {
		case msg := <-ch:
			var entry models.QueueEntry
			if err := json.Unmarshal([]byte(msg.Payload), &entry); err != nil {
				log.Printf("Error unmarshaling queue update: %v", err)
				continue
			}
			callback(&entry)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// UpdateQueueCache updates queue entry in Redis cache
func (rs *RealtimeService) UpdateQueueCache(ctx context.Context, entry *models.QueueEntry) error {
	key := fmt.Sprintf("queue:entry:%s", entry.ID)
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return rs.redis.Set(ctx, key, data, 1*time.Hour).Err()
}

// GetQueueCache retrieves queue entry from Redis cache
func (rs *RealtimeService) GetQueueCache(ctx context.Context, entryID string) (*models.QueueEntry, error) {
	key := fmt.Sprintf("queue:entry:%s", entryID)
	data, err := rs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var entry models.QueueEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// InvalidateQueueCache removes queue entry from cache
func (rs *RealtimeService) InvalidateQueueCache(ctx context.Context, entryID string) error {
	key := fmt.Sprintf("queue:entry:%s", entryID)
	return rs.redis.Del(ctx, key).Err()
}

// SetActiveQueueSnapshot stores current active queue state
func (rs *RealtimeService) SetActiveQueueSnapshot(ctx context.Context, entries []models.QueueEntry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	key := "queue:active:snapshot"
	return rs.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

// GetActiveQueueSnapshot retrieves active queue snapshot
func (rs *RealtimeService) GetActiveQueueSnapshot(ctx context.Context) ([]models.QueueEntry, error) {
	key := "queue:active:snapshot"
	data, err := rs.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var entries []models.QueueEntry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// IncrementTokenCounter increments daily token counter atomically
func (rs *RealtimeService) IncrementTokenCounter(ctx context.Context, date string) (int64, error) {
	key := fmt.Sprintf("queue:token:counter:%s", date)
	val, err := rs.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry to 48 hours
	rs.redis.Expire(ctx, key, 48*time.Hour)

	return val, nil
}

// GetCurrentQueueLength gets current queue length from Redis
func (rs *RealtimeService) GetCurrentQueueLength(ctx context.Context) (int64, error) {
	key := "queue:length"
	val, err := rs.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// UpdateQueueLength updates current queue length
func (rs *RealtimeService) UpdateQueueLength(ctx context.Context, length int64) error {
	key := "queue:length"
	return rs.redis.Set(ctx, key, length, 1*time.Hour).Err()
}
