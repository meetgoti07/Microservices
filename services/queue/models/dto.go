package models

import "time"

// CreateQueueEntryRequest represents request to create queue entry
type CreateQueueEntryRequest struct {
	OrderID         string `json:"order_id" binding:"required"`
	UserID          string `json:"user_id" binding:"required"`
	UserName        string `json:"user_name"`
	UserPhone       string `json:"user_phone"`
	TokenType       string `json:"token_type"`
	Priority        string `json:"priority"`
	IsExpressQueue  bool   `json:"is_express_queue"`
	SpecialHandling string `json:"special_handling"`
	ItemCount       int    `json:"item_count"`
}

// UpdateQueueStatusRequest represents request to update queue status
type UpdateQueueStatusRequest struct {
	Status          string  `json:"status" binding:"required"`
	AssignedCounter *string `json:"assigned_counter"`
	AssignedStaff   *string `json:"assigned_staff"`
	Notes           *string `json:"notes"`
	Reason          *string `json:"reason"`
}

// UpdateQueuePriorityRequest represents request to update priority
type UpdateQueuePriorityRequest struct {
	Priority string  `json:"priority" binding:"required"`
	Reason   *string `json:"reason"`
}

// AssignStaffRequest represents request to assign staff
type AssignStaffRequest struct {
	StaffID   string  `json:"staff_id" binding:"required"`
	StaffName string  `json:"staff_name"`
	Counter   *string `json:"counter"`
}

// QueuePositionResponse represents queue position info
type QueuePositionResponse struct {
	QueueEntry        *QueueEntry `json:"queue_entry"`
	Position          int         `json:"position"`
	EstimatedWaitTime int         `json:"estimated_wait_time"`
	EstimatedReadyTime *time.Time `json:"estimated_ready_time,omitempty"`
	PeopleAhead       int         `json:"people_ahead"`
}

// CurrentQueueResponse represents current queue state
type CurrentQueueResponse struct {
	Waiting     []QueueEntry `json:"waiting"`
	InProgress  []QueueEntry `json:"in_progress"`
	Ready       []QueueEntry `json:"ready"`
	TotalActive int          `json:"total_active"`
}

// QueueStatsResponse represents queue statistics
type QueueStatsResponse struct {
	Date                 string  `json:"date"`
	TotalInQueue         int     `json:"total_in_queue"`
	WaitingCount         int     `json:"waiting_count"`
	InProgressCount      int     `json:"in_progress_count"`
	ReadyCount           int     `json:"ready_count"`
	CompletedToday       int     `json:"completed_today"`
	CancelledToday       int     `json:"cancelled_today"`
	AvgWaitTime          int     `json:"avg_wait_time"`
	AvgPreparationTime   int     `json:"avg_preparation_time"`
	CurrentLoad          float64 `json:"current_load"`
	OnTimeCompletionRate float64 `json:"on_time_completion_rate"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
