package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gin-quickstart/config"
	"gin-quickstart/models"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(cfg *config.Config) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(cfg.KafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	log.Println("Kafka producer created successfully")
	return &KafkaProducer{producer: producer}, nil
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

// PublishQueuePositionUpdate publishes position update event
func (kp *KafkaProducer) PublishQueuePositionUpdate(entry *models.QueueEntry) error {
	event := map[string]interface{}{
		"event_type":          "queue.position.updated",
		"queue_entry_id":      entry.ID,
		"order_id":            entry.OrderID,
		"user_id":             entry.UserID,
		"token_number":        entry.TokenNumber,
		"position":            entry.Position,
		"estimated_wait_time": entry.EstimatedWaitTime,
		"estimated_ready_time": entry.EstimatedReadyTime,
		"status":              entry.Status,
		"timestamp":           time.Now().UTC(),
	}

	return kp.publishEvent("queue.events", event)
}

// PublishQueueStatusChanged publishes status change event
func (kp *KafkaProducer) PublishQueueStatusChanged(entry *models.QueueEntry, oldStatus, newStatus string) error {
	event := map[string]interface{}{
		"event_type":          "queue.status.changed",
		"queue_entry_id":      entry.ID,
		"order_id":            entry.OrderID,
		"user_id":             entry.UserID,
		"token_number":        entry.TokenNumber,
		"old_status":          oldStatus,
		"new_status":          newStatus,
		"position":            entry.Position,
		"estimated_wait_time": entry.EstimatedWaitTime,
		"timestamp":           time.Now().UTC(),
	}

	return kp.publishEvent("queue.events", event)
}

// PublishQueueAlmostReady publishes almost ready notification
func (kp *KafkaProducer) PublishQueueAlmostReady(entry *models.QueueEntry) error {
	event := map[string]interface{}{
		"event_type":          "queue.almost.ready",
		"queue_entry_id":      entry.ID,
		"order_id":            entry.OrderID,
		"user_id":             entry.UserID,
		"token_number":        entry.TokenNumber,
		"position":            entry.Position,
		"estimated_wait_time": entry.EstimatedWaitTime,
		"timestamp":           time.Now().UTC(),
		"notification_type":   "ALMOST_READY",
	}

	return kp.publishEvent("notification.events", event)
}

// PublishQueueReady publishes ready notification
func (kp *KafkaProducer) PublishQueueReady(entry *models.QueueEntry) error {
	event := map[string]interface{}{
		"event_type":     "queue.ready",
		"queue_entry_id": entry.ID,
		"order_id":       entry.OrderID,
		"user_id":        entry.UserID,
		"token_number":   entry.TokenNumber,
		"timestamp":      time.Now().UTC(),
		"notification_type": "READY",
	}

	return kp.publishEvent("notification.events", event)
}

// PublishQueueCompleted publishes completion event
func (kp *KafkaProducer) PublishQueueCompleted(entry *models.QueueEntry) error {
	event := map[string]interface{}{
		"event_type":     "queue.completed",
		"queue_entry_id": entry.ID,
		"order_id":       entry.OrderID,
		"user_id":        entry.UserID,
		"token_number":   entry.TokenNumber,
		"timestamp":      time.Now().UTC(),
	}

	return kp.publishEvent("queue.events", event)
}

// PublishQueueAdvanced publishes queue advance event
func (kp *KafkaProducer) PublishQueueAdvanced(entry *models.QueueEntry) error {
	event := map[string]interface{}{
		"event_type":     "queue.advanced",
		"queue_entry_id": entry.ID,
		"order_id":       entry.OrderID,
		"token_number":   entry.TokenNumber,
		"new_status":     entry.Status,
		"timestamp":      time.Now().UTC(),
	}

	return kp.publishEvent("queue.events", event)
}

func (kp *KafkaProducer) publishEvent(topic string, event map[string]interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(fmt.Sprintf("%v", event["queue_entry_id"])),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Published event to %s: partition=%d, offset=%d, event_type=%s",
		topic, partition, offset, event["event_type"])

	return nil
}
