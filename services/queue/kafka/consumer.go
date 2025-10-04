package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gin-quickstart/config"
	"gin-quickstart/models"
	"gin-quickstart/services"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer      sarama.ConsumerGroup
	queueService  *services.QueueService
	topics        []string
	ready         chan bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// OrderCreatedEvent represents order creation event from Order Service
type OrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	UserPhone   string    `json:"user_phone"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64   `json:"total_amount"`
	Priority    string    `json:"priority,omitempty"`
	IsExpress   bool      `json:"is_express,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderItem struct {
	MenuItemID string  `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

// OrderStatusEvent represents order status change event
type OrderStatusEvent struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewKafkaConsumer(cfg *config.Config, queueService *services.QueueService) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	ctx, cancel := context.WithCancel(context.Background())

	consumer, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumer:     consumer,
		queueService: queueService,
		topics:       []string{"order.created", "order.status.changed"},
		ready:        make(chan bool),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (kc *KafkaConsumer) Start() error {
	go func() {
		for {
			select {
			case <-kc.ctx.Done():
				return
			default:
				if err := kc.consumer.Consume(kc.ctx, kc.topics, kc); err != nil {
					log.Printf("Error from consumer: %v", err)
					time.Sleep(5 * time.Second) // Backoff before retry
				}
			}
		}
	}()

	// Wait for consumer to be ready
	<-kc.ready
	log.Println("Kafka consumer started and ready")
	
	return nil
}

func (kc *KafkaConsumer) Stop() error {
	kc.cancel()
	return kc.consumer.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (kc *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(kc.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (kc *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			log.Printf("Message received: topic=%s, partition=%d, offset=%d", 
				message.Topic, message.Partition, message.Offset)

			if err := kc.handleMessage(message); err != nil {
				log.Printf("Error handling message: %v", err)
				// Continue processing other messages even if one fails
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (kc *KafkaConsumer) handleMessage(message *sarama.ConsumerMessage) error {
	ctx := context.Background()

	switch message.Topic {
	case "order.created":
		return kc.handleOrderCreated(ctx, message.Value)
	case "order.status.changed":
		return kc.handleOrderStatusChanged(ctx, message.Value)
	default:
		log.Printf("Unknown topic: %s", message.Topic)
		return nil
	}
}

func (kc *KafkaConsumer) handleOrderCreated(ctx context.Context, data []byte) error {
	var event OrderCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	log.Printf("Processing order created event: order_id=%s, user_id=%s", event.OrderID, event.UserID)

	// Check if queue entry already exists
	existing, _ := kc.queueService.GetQueueEntryByOrderID(ctx, event.OrderID)
	if existing != nil {
		log.Printf("Queue entry already exists for order %s", event.OrderID)
		return nil
	}

	// Determine priority based on order
	priority := event.Priority
	if priority == "" {
		priority = "NORMAL"
	}

	// Determine if express queue
	isExpress := event.IsExpress
	itemCount := 0
	for _, item := range event.Items {
		itemCount += item.Quantity
	}

	// Auto-qualify for express if <= 3 items
	if itemCount <= 3 && !isExpress {
		isExpress = true
		priority = "HIGH"
	}

	// Create queue entry
	req := &models.CreateQueueEntryRequest{
		OrderID:        event.OrderID,
		UserID:         event.UserID,
		UserName:       event.UserName,
		UserPhone:      event.UserPhone,
		TokenType:      determineTokenType(itemCount, isExpress),
		Priority:       priority,
		IsExpressQueue: isExpress,
		ItemCount:      itemCount,
	}

	entry, err := kc.queueService.CreateQueueEntry(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create queue entry: %w", err)
	}

	log.Printf("Queue entry created: token=%s, position=%d, estimated_wait=%d mins",
		entry.TokenNumber, entry.Position, entry.EstimatedWaitTime)

	// Publish queue entry created event
	go kc.publishQueueEntryCreated(entry)

	return nil
}

func (kc *KafkaConsumer) handleOrderStatusChanged(ctx context.Context, data []byte) error {
	var event OrderStatusEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal order status event: %w", err)
	}

	log.Printf("Processing order status changed: order_id=%s, status=%s", event.OrderID, event.Status)

	// Get queue entry for order
	entry, err := kc.queueService.GetQueueEntryByOrderID(ctx, event.OrderID)
	if err != nil {
		log.Printf("Queue entry not found for order %s", event.OrderID)
		return nil
	}

	// Map order status to queue status
	queueStatus := mapOrderStatusToQueueStatus(event.Status)
	if queueStatus == "" {
		log.Printf("No queue status mapping for order status: %s", event.Status)
		return nil
	}

	// Update queue status
	req := &models.UpdateQueueStatusRequest{
		Status: queueStatus,
	}

	if err := kc.queueService.UpdateQueueStatus(ctx, entry.ID, req, "system", "System"); err != nil {
		return fmt.Errorf("failed to update queue status: %w", err)
	}

	log.Printf("Queue status updated: token=%s, status=%s", entry.TokenNumber, queueStatus)

	return nil
}

func (kc *KafkaConsumer) publishQueueEntryCreated(entry *models.QueueEntry) {
	// Publish to notification service via Kafka
	event := map[string]interface{}{
		"event_type":          "queue.entry.created",
		"queue_entry_id":      entry.ID,
		"order_id":            entry.OrderID,
		"user_id":             entry.UserID,
		"token_number":        entry.TokenNumber,
		"position":            entry.Position,
		"estimated_wait_time": entry.EstimatedWaitTime,
		"estimated_ready_time": entry.EstimatedReadyTime,
		"created_at":          entry.CreatedAt,
	}

	data, _ := json.Marshal(event)
	
	// Send to Kafka topic for notifications
	producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, nil)
	if err != nil {
		log.Printf("Failed to create producer: %v", err)
		return
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: "queue.events",
		Value: sarama.ByteEncoder(data),
	}

	if _, _, err := producer.SendMessage(msg); err != nil {
		log.Printf("Failed to publish queue entry created event: %v", err)
	} else {
		log.Printf("Published queue entry created event: token=%s", entry.TokenNumber)
	}
}

func determineTokenType(itemCount int, isExpress bool) string {
	if isExpress {
		return "EXPRESS"
	}
	if itemCount > 10 {
		return "BULK"
	}
	return "REGULAR"
}

func mapOrderStatusToQueueStatus(orderStatus string) string {
	statusMap := map[string]string{
		"CONFIRMED":  "WAITING",
		"PREPARING":  "IN_PROGRESS",
		"READY":      "READY",
		"COMPLETED":  "COMPLETED",
		"CANCELLED":  "CANCELLED",
		"FAILED":     "CANCELLED",
	}
	return statusMap[orderStatus]
}
