package models

import (
	"time"
)

// QueueEntry represents a queue entry in the system
type QueueEntry struct {
	ID                        string     `gorm:"column:id;primaryKey" json:"id"`
	OrderID                   string     `gorm:"column:order_id;uniqueIndex;not null" json:"order_id"`
	UserID                    string     `gorm:"column:user_id;index;not null" json:"user_id"`
	UserName                  *string    `gorm:"column:user_name" json:"user_name,omitempty"`
	UserPhone                 *string    `gorm:"column:user_phone" json:"user_phone,omitempty"`
	TokenNumber               string     `gorm:"column:token_number;uniqueIndex;not null" json:"token_number"`
	TokenType                 string     `gorm:"column:token_type;type:ENUM('REGULAR','EXPRESS','BULK','SPECIAL','STAFF');default:'REGULAR'" json:"token_type"`
	Status                    string     `gorm:"column:status;type:ENUM('WAITING','IN_PROGRESS','READY','COMPLETED','CANCELLED','NO_SHOW','EXPIRED');default:'WAITING';index" json:"status"`
	Priority                  string     `gorm:"column:priority;type:ENUM('LOW','NORMAL','HIGH','URGENT','VIP');default:'NORMAL';index" json:"priority"`
	Position                  int        `gorm:"column:position;not null;index" json:"position"`
	EstimatedWaitTime         int        `gorm:"column:estimated_wait_time;default:0" json:"estimated_wait_time"`
	EstimatedReadyTime        *time.Time `gorm:"column:estimated_ready_time;index" json:"estimated_ready_time,omitempty"`
	ActualStartTime           *time.Time `gorm:"column:actual_start_time" json:"actual_start_time,omitempty"`
	ActualReadyTime           *time.Time `gorm:"column:actual_ready_time" json:"actual_ready_time,omitempty"`
	ActualCompletionTime      *time.Time `gorm:"column:actual_completion_time" json:"actual_completion_time,omitempty"`
	AssignedCounter           *string    `gorm:"column:assigned_counter;index" json:"assigned_counter,omitempty"`
	AssignedStaff             *string    `gorm:"column:assigned_staff;index" json:"assigned_staff,omitempty"`
	AssignedStaffName         *string    `gorm:"column:assigned_staff_name" json:"assigned_staff_name,omitempty"`
	AverageItemPreparationTime *int      `gorm:"column:average_item_preparation_time" json:"average_item_preparation_time,omitempty"`
	IsExpressQueue            bool       `gorm:"column:is_express_queue;default:false" json:"is_express_queue"`
	SpecialHandling           *string    `gorm:"column:special_handling" json:"special_handling,omitempty"`
	Notes                     *string    `gorm:"column:notes" json:"notes,omitempty"`
	CreatedAt                 time.Time  `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt                 time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (QueueEntry) TableName() string {
	return "queue_entries"
}

// QueueNotificationSent tracks notifications sent for queue entries
type QueueNotificationSent struct {
	ID               string    `gorm:"column:id;primaryKey" json:"id"`
	QueueEntryID     string    `gorm:"column:queue_entry_id;index;not null" json:"queue_entry_id"`
	NotificationType string    `gorm:"column:notification_type;type:ENUM('ORDER_CONFIRMED','POSITION_UPDATE','ALMOST_READY','READY','REMINDER');not null;index" json:"notification_type"`
	Channel          string    `gorm:"column:channel;type:ENUM('PUSH','IN_APP','SMS','EMAIL');not null" json:"channel"`
	SentAt           time.Time `gorm:"column:sent_at;index" json:"sent_at"`
}

func (QueueNotificationSent) TableName() string {
	return "queue_notifications_sent"
}

// QueuePositionHistory tracks position changes
type QueuePositionHistory struct {
	ID                  string     `gorm:"column:id;primaryKey" json:"id"`
	QueueEntryID        string     `gorm:"column:queue_entry_id;index;not null" json:"queue_entry_id"`
	OldPosition         int        `gorm:"column:old_position;not null" json:"old_position"`
	NewPosition         int        `gorm:"column:new_position;not null" json:"new_position"`
	OldStatus           string     `gorm:"column:old_status;not null" json:"old_status"`
	NewStatus           string     `gorm:"column:new_status;not null" json:"new_status"`
	EstimatedWaitTime   *int       `gorm:"column:estimated_wait_time" json:"estimated_wait_time,omitempty"`
	EstimatedReadyTime  *time.Time `gorm:"column:estimated_ready_time" json:"estimated_ready_time,omitempty"`
	Reason              *string    `gorm:"column:reason" json:"reason,omitempty"`
	Timestamp           time.Time  `gorm:"column:timestamp;index" json:"timestamp"`
}

func (QueuePositionHistory) TableName() string {
	return "queue_position_history"
}

// QueueConfiguration holds queue settings
type QueueConfiguration struct {
	ID                              string    `gorm:"column:id;primaryKey" json:"id"`
	MaxConcurrentOrders             int       `gorm:"column:max_concurrent_orders;default:10" json:"max_concurrent_orders"`
	AvgPreparationTimePerItem       int       `gorm:"column:avg_preparation_time_per_item;default:5" json:"avg_preparation_time_per_item"`
	BufferTime                      int       `gorm:"column:buffer_time;default:2" json:"buffer_time"`
	ExpressQueueEnabled             bool      `gorm:"column:express_queue_enabled;default:false" json:"express_queue_enabled"`
	ExpressQueueMaxItems            int       `gorm:"column:express_queue_max_items;default:3" json:"express_queue_max_items"`
	MaxWaitTimeAlert                int       `gorm:"column:max_wait_time_alert;default:30" json:"max_wait_time_alert"`
	TokenExpiryTime                 int       `gorm:"column:token_expiry_time;default:60" json:"token_expiry_time"`
	AutoNotificationEnabled         bool      `gorm:"column:auto_notification_enabled;default:true" json:"auto_notification_enabled"`
	NotificationPositionThreshold   int       `gorm:"column:notification_position_threshold;default:5" json:"notification_position_threshold"`
	NotificationAlmostReadyThreshold int      `gorm:"column:notification_almost_ready_threshold;default:2" json:"notification_almost_ready_threshold"`
	UpdatedAt                       time.Time `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy                       *string   `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (QueueConfiguration) TableName() string {
	return "queue_configuration"
}

// QueueWorkingHours defines operating hours
type QueueWorkingHours struct {
	ID              string `gorm:"column:id;primaryKey" json:"id"`
	ConfigurationID string `gorm:"column:configuration_id;index;not null" json:"configuration_id"`
	Day             string `gorm:"column:day;type:ENUM('MONDAY','TUESDAY','WEDNESDAY','THURSDAY','FRIDAY','SATURDAY','SUNDAY');not null" json:"day"`
	OpenTime        string `gorm:"column:open_time;not null" json:"open_time"`
	CloseTime       string `gorm:"column:close_time;not null" json:"close_time"`
	IsOpen          bool   `gorm:"column:is_open;default:true" json:"is_open"`
}

func (QueueWorkingHours) TableName() string {
	return "queue_working_hours"
}

// QueuePriorityMultiplier defines priority time multipliers
type QueuePriorityMultiplier struct {
	ID              string  `gorm:"column:id;primaryKey" json:"id"`
	ConfigurationID string  `gorm:"column:configuration_id;index;not null" json:"configuration_id"`
	Priority        string  `gorm:"column:priority;type:ENUM('LOW','NORMAL','HIGH','URGENT','VIP');not null" json:"priority"`
	Multiplier      float64 `gorm:"column:multiplier;default:1.00" json:"multiplier"`
}

func (QueuePriorityMultiplier) TableName() string {
	return "queue_priority_multipliers"
}

// QueueDisplayAnnouncement for display announcements
type QueueDisplayAnnouncement struct {
	ID           string     `gorm:"column:id;primaryKey" json:"id"`
	Message      string     `gorm:"column:message;not null" json:"message"`
	Type         string     `gorm:"column:type;type:ENUM('INFO','WARNING','URGENT');default:'INFO'" json:"type"`
	Priority     int        `gorm:"column:priority;default:0;index" json:"priority"`
	IsActive     bool       `gorm:"column:is_active;default:true;index" json:"is_active"`
	DisplayUntil *time.Time `gorm:"column:display_until;index" json:"display_until,omitempty"`
	CreatedBy    *string    `gorm:"column:created_by" json:"created_by,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (QueueDisplayAnnouncement) TableName() string {
	return "queue_display_announcements"
}

// StaffQueueActionLog logs staff actions
type StaffQueueActionLog struct {
	ID              string     `gorm:"column:id;primaryKey" json:"id"`
	QueueEntryID    string     `gorm:"column:queue_entry_id;index;not null" json:"queue_entry_id"`
	StaffID         string     `gorm:"column:staff_id;index;not null" json:"staff_id"`
	StaffName       *string    `gorm:"column:staff_name" json:"staff_name,omitempty"`
	Action          string     `gorm:"column:action;type:ENUM('START_PREPARATION','MARK_READY','MARK_COMPLETED','CANCEL','REASSIGN','ADJUST_PRIORITY','ADD_NOTE');not null;index" json:"action"`
	OldStatus       *string    `gorm:"column:old_status" json:"old_status,omitempty"`
	NewStatus       *string    `gorm:"column:new_status" json:"new_status,omitempty"`
	OldPriority     *string    `gorm:"column:old_priority" json:"old_priority,omitempty"`
	NewPriority     *string    `gorm:"column:new_priority" json:"new_priority,omitempty"`
	AssignedCounter *string    `gorm:"column:assigned_counter" json:"assigned_counter,omitempty"`
	AssignedStaff   *string    `gorm:"column:assigned_staff" json:"assigned_staff,omitempty"`
	Note            *string    `gorm:"column:note" json:"note,omitempty"`
	Reason          *string    `gorm:"column:reason" json:"reason,omitempty"`
	Timestamp       time.Time  `gorm:"column:timestamp;index" json:"timestamp"`
}

func (StaffQueueActionLog) TableName() string {
	return "staff_queue_actions_log"
}

// QueueStatistics holds daily statistics
type QueueStatistics struct {
	ID                    string    `gorm:"column:id;primaryKey" json:"id"`
	Date                  time.Time `gorm:"column:date;uniqueIndex;not null" json:"date"`
	TotalInQueue          int       `gorm:"column:total_in_queue;default:0" json:"total_in_queue"`
	WaitingCount          int       `gorm:"column:waiting_count;default:0" json:"waiting_count"`
	InProgressCount       int       `gorm:"column:in_progress_count;default:0" json:"in_progress_count"`
	ReadyCount            int       `gorm:"column:ready_count;default:0" json:"ready_count"`
	CompletedToday        int       `gorm:"column:completed_today;default:0" json:"completed_today"`
	CancelledToday        int       `gorm:"column:cancelled_today;default:0" json:"cancelled_today"`
	NoShowToday           int       `gorm:"column:no_show_today;default:0" json:"no_show_today"`
	ExpiredToday          int       `gorm:"column:expired_today;default:0" json:"expired_today"`
	AvgWaitTime           int       `gorm:"column:avg_wait_time;default:0" json:"avg_wait_time"`
	AvgPreparationTime    int       `gorm:"column:avg_preparation_time;default:0" json:"avg_preparation_time"`
	LongestWaitTime       int       `gorm:"column:longest_wait_time;default:0" json:"longest_wait_time"`
	ShortestWaitTime      int       `gorm:"column:shortest_wait_time;default:0" json:"shortest_wait_time"`
	CurrentLoad           float64   `gorm:"column:current_load;default:0.00" json:"current_load"`
	PeakLoad              float64   `gorm:"column:peak_load;default:0.00" json:"peak_load"`
	PeakLoadTime          *string   `gorm:"column:peak_load_time" json:"peak_load_time,omitempty"`
	OnTimeCompletionRate  float64   `gorm:"column:on_time_completion_rate;default:0.00" json:"on_time_completion_rate"`
	NoShowRate            float64   `gorm:"column:no_show_rate;default:0.00" json:"no_show_rate"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (QueueStatistics) TableName() string {
	return "queue_statistics"
}

// QueueHourlyStatistics holds hourly statistics
type QueueHourlyStatistics struct {
	ID                  string    `gorm:"column:id;primaryKey" json:"id"`
	Date                time.Time `gorm:"column:date;not null" json:"date"`
	Hour                int       `gorm:"column:hour;not null" json:"hour"`
	OrderCount          int       `gorm:"column:order_count;default:0" json:"order_count"`
	AvgWaitTime         int       `gorm:"column:avg_wait_time;default:0" json:"avg_wait_time"`
	AvgPreparationTime  int       `gorm:"column:avg_preparation_time;default:0" json:"avg_preparation_time"`
	CompletedCount      int       `gorm:"column:completed_count;default:0" json:"completed_count"`
	CancelledCount      int       `gorm:"column:cancelled_count;default:0" json:"cancelled_count"`
	PeakPosition        int       `gorm:"column:peak_position;default:0" json:"peak_position"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (QueueHourlyStatistics) TableName() string {
	return "queue_hourly_statistics"
}

// QueueTokenCounter tracks token generation
type QueueTokenCounter struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	Date          time.Time `gorm:"column:date;uniqueIndex;not null" json:"date"`
	CurrentNumber int       `gorm:"column:current_number;default:0" json:"current_number"`
	Prefix        string    `gorm:"column:prefix;default:'A'" json:"prefix"`
	LastResetAt   time.Time `gorm:"column:last_reset_at" json:"last_reset_at"`
}

func (QueueTokenCounter) TableName() string {
	return "queue_token_counter"
}
