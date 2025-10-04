package services

import (
	"context"
	"errors"
	"time"

	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/utils"

	"gorm.io/gorm"
)

type QueueService struct {
	db *gorm.DB
}

func NewQueueService() *QueueService {
	return &QueueService{
		db: database.GetDB(),
	}
}

// CreateQueueEntry creates a new queue entry
func (s *QueueService) CreateQueueEntry(ctx context.Context, req *models.CreateQueueEntryRequest) (*models.QueueEntry, error) {
	// Check if order already in queue
	var existing models.QueueEntry
	if err := s.db.Where("order_id = ?", req.OrderID).First(&existing).Error; err == nil {
		return nil, errors.New("order already in queue")
	}

	// Get configuration
	config, err := s.GetConfiguration(ctx)
	if err != nil {
		return nil, err
	}

	// Generate token number
	tokenNumber, err := utils.GenerateTokenNumber(s.db)
	if err != nil {
		return nil, err
	}

	// Calculate position
	var currentMaxPosition int
	s.db.Model(&models.QueueEntry{}).
		Where("status IN ?", []string{"WAITING", "IN_PROGRESS"}).
		Select("COALESCE(MAX(position), 0)").
		Scan(&currentMaxPosition)

	newPosition := currentMaxPosition + 1

	// Set defaults
	tokenType := req.TokenType
	if tokenType == "" {
		tokenType = "REGULAR"
	}

	priority := req.Priority
	if priority == "" {
		priority = "NORMAL"
	}

	// Calculate estimated times
	estimatedWaitTime := utils.CalculateEstimatedWaitTime(
		newPosition,
		config.AvgPreparationTimePerItem,
		config.BufferTime,
	)
	estimatedReadyTime := utils.CalculateEstimatedReadyTime(estimatedWaitTime)

	// Create entry
	entry := &models.QueueEntry{
		ID:                         utils.GenerateUUID(),
		OrderID:                    req.OrderID,
		UserID:                     req.UserID,
		UserName:                   utils.StringPtr(req.UserName),
		UserPhone:                  utils.StringPtr(req.UserPhone),
		TokenNumber:                tokenNumber,
		TokenType:                  tokenType,
		Status:                     "WAITING",
		Priority:                   priority,
		Position:                   newPosition,
		EstimatedWaitTime:          estimatedWaitTime,
		EstimatedReadyTime:         &estimatedReadyTime,
		IsExpressQueue:             req.IsExpressQueue,
		SpecialHandling:            utils.StringPtr(req.SpecialHandling),
		AverageItemPreparationTime: utils.IntPtr(config.AvgPreparationTimePerItem * req.ItemCount),
		CreatedAt:                  time.Now().UTC(),
		UpdatedAt:                  time.Now().UTC(),
	}

	if err := s.db.Create(entry).Error; err != nil {
		return nil, err
	}

	// Cache in Redis
	utils.CacheQueueEntry(ctx, entry)

	// Update statistics
	go s.UpdateStatistics(ctx)

	return entry, nil
}

// GetQueueEntryByToken retrieves queue entry by token number
func (s *QueueService) GetQueueEntryByToken(ctx context.Context, token string) (*models.QueueEntry, error) {
	var entry models.QueueEntry
	if err := s.db.Where("token_number = ?", token).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetQueueEntryByID retrieves queue entry by ID
func (s *QueueService) GetQueueEntryByID(ctx context.Context, id string) (*models.QueueEntry, error) {
	var entry models.QueueEntry
	if err := s.db.Where("id = ?", id).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetQueueEntryByOrderID retrieves queue entry by order ID
func (s *QueueService) GetQueueEntryByOrderID(ctx context.Context, orderID string) (*models.QueueEntry, error) {
	var entry models.QueueEntry
	if err := s.db.Where("order_id = ?", orderID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetQueuePosition gets position info for a token
func (s *QueueService) GetQueuePosition(ctx context.Context, token string) (*models.QueuePositionResponse, error) {
	entry, err := s.GetQueueEntryByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Count people ahead
	var peopleAhead int64
	s.db.Model(&models.QueueEntry{}).
		Where("status IN ? AND position < ?", []string{"WAITING", "IN_PROGRESS"}, entry.Position).
		Count(&peopleAhead)

	return &models.QueuePositionResponse{
		QueueEntry:         entry,
		Position:           entry.Position,
		EstimatedWaitTime:  entry.EstimatedWaitTime,
		EstimatedReadyTime: entry.EstimatedReadyTime,
		PeopleAhead:        int(peopleAhead),
	}, nil
}

// GetCurrentQueue gets current queue state
func (s *QueueService) GetCurrentQueue(ctx context.Context) (*models.CurrentQueueResponse, error) {
	var waiting, inProgress, ready []models.QueueEntry

	s.db.Where("status = ?", "WAITING").Order("position ASC").Find(&waiting)
	s.db.Where("status = ?", "IN_PROGRESS").Order("position ASC").Find(&inProgress)
	s.db.Where("status = ?", "READY").Order("actual_ready_time DESC").Limit(20).Find(&ready)

	return &models.CurrentQueueResponse{
		Waiting:     waiting,
		InProgress:  inProgress,
		Ready:       ready,
		TotalActive: len(waiting) + len(inProgress) + len(ready),
	}, nil
}

// UpdateQueueStatus updates queue entry status
func (s *QueueService) UpdateQueueStatus(ctx context.Context, entryID string, req *models.UpdateQueueStatusRequest, staffID string, staffName string) error {
	var entry models.QueueEntry
	if err := s.db.Where("id = ?", entryID).First(&entry).Error; err != nil {
		return err
	}

	oldStatus := entry.Status
	oldPosition := entry.Position

	// Update status
	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now().UTC(),
	}

	// Set timestamps based on status
	now := time.Now().UTC()
	switch req.Status {
	case "IN_PROGRESS":
		if entry.ActualStartTime == nil {
			updates["actual_start_time"] = now
		}
		if req.AssignedCounter != nil {
			updates["assigned_counter"] = *req.AssignedCounter
		}
		if req.AssignedStaff != nil {
			updates["assigned_staff"] = *req.AssignedStaff
		}
	case "READY":
		if entry.ActualReadyTime == nil {
			updates["actual_ready_time"] = now
		}
	case "COMPLETED":
		if entry.ActualCompletionTime == nil {
			updates["actual_completion_time"] = now
		}
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if err := s.db.Model(&entry).Updates(updates).Error; err != nil {
		return err
	}

	// Log action
	s.LogStaffAction(ctx, entryID, staffID, staffName, "MARK_"+req.Status, &oldStatus, &req.Status, nil, nil, req.Reason)

	// Record position history
	s.RecordPositionHistory(ctx, entryID, oldPosition, entry.Position, oldStatus, req.Status, req.Reason)

	// Invalidate cache
	utils.InvalidateQueueCache(ctx, entryID)

	// Recalculate positions if needed
	if req.Status == "COMPLETED" || req.Status == "CANCELLED" || req.Status == "NO_SHOW" {
		go s.RecalculatePositions(ctx)
	}

	// Update statistics
	go s.UpdateStatistics(ctx)

	return nil
}

// UpdateQueuePriority updates queue entry priority
func (s *QueueService) UpdateQueuePriority(ctx context.Context, entryID string, req *models.UpdateQueuePriorityRequest, staffID string, staffName string) error {
	var entry models.QueueEntry
	if err := s.db.Where("id = ?", entryID).First(&entry).Error; err != nil {
		return err
	}

	oldPriority := entry.Priority

	updates := map[string]interface{}{
		"priority":   req.Priority,
		"updated_at": time.Now().UTC(),
	}

	if err := s.db.Model(&entry).Updates(updates).Error; err != nil {
		return err
	}

	// Log action
	s.LogStaffAction(ctx, entryID, staffID, staffName, "ADJUST_PRIORITY", nil, nil, &oldPriority, &req.Priority, req.Reason)

	// Invalidate cache
	utils.InvalidateQueueCache(ctx, entryID)

	// Recalculate wait times
	go s.RecalculatePositions(ctx)

	return nil
}

// AssignStaff assigns staff to queue entry
func (s *QueueService) AssignStaff(ctx context.Context, entryID string, req *models.AssignStaffRequest, staffID string, staffName string) error {
	updates := map[string]interface{}{
		"assigned_staff":      req.StaffID,
		"assigned_staff_name": req.StaffName,
		"updated_at":          time.Now().UTC(),
	}

	if req.Counter != nil {
		updates["assigned_counter"] = *req.Counter
	}

	if err := s.db.Model(&models.QueueEntry{}).Where("id = ?", entryID).Updates(updates).Error; err != nil {
		return err
	}

	// Log action
	s.LogStaffAction(ctx, entryID, staffID, staffName, "REASSIGN", nil, nil, nil, nil, utils.StringPtr("Staff assigned"))

	// Invalidate cache
	utils.InvalidateQueueCache(ctx, entryID)

	return nil
}

// AdvanceQueue advances the queue (staff action)
func (s *QueueService) AdvanceQueue(ctx context.Context, staffID string, staffName string) error {
	// Get next waiting entry
	var entry models.QueueEntry
	if err := s.db.Where("status = ?", "WAITING").
		Order("priority DESC, position ASC").
		First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no entries in queue")
		}
		return err
	}

	// Move to IN_PROGRESS
	req := &models.UpdateQueueStatusRequest{
		Status: "IN_PROGRESS",
	}

	return s.UpdateQueueStatus(ctx, entry.ID, req, staffID, staffName)
}

// RecalculatePositions recalculates all positions and estimated times
func (s *QueueService) RecalculatePositions(ctx context.Context) error {
	var entries []models.QueueEntry
	if err := s.db.Where("status IN ?", []string{"WAITING", "IN_PROGRESS"}).
		Order("priority DESC, position ASC").
		Find(&entries).Error; err != nil {
		return err
	}

	config, err := s.GetConfiguration(ctx)
	if err != nil {
		return err
	}

	for i, entry := range entries {
		newPosition := i + 1
		estimatedWaitTime := utils.CalculateEstimatedWaitTime(newPosition, config.AvgPreparationTimePerItem, config.BufferTime)
		estimatedReadyTime := utils.CalculateEstimatedReadyTime(estimatedWaitTime)

		s.db.Model(&models.QueueEntry{}).Where("id = ?", entry.ID).Updates(map[string]interface{}{
			"position":              newPosition,
			"estimated_wait_time":   estimatedWaitTime,
			"estimated_ready_time":  estimatedReadyTime,
			"updated_at":            time.Now().UTC(),
		})
	}

	return nil
}

// GetConfiguration gets queue configuration
func (s *QueueService) GetConfiguration(ctx context.Context) (*models.QueueConfiguration, error) {
	var config models.QueueConfiguration
	if err := s.db.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateConfiguration updates queue configuration
func (s *QueueService) UpdateConfiguration(ctx context.Context, config *models.QueueConfiguration, userID string) error {
	config.UpdatedAt = time.Now().UTC()
	config.UpdatedBy = &userID
	
	if err := s.db.Save(config).Error; err != nil {
		return err
	}
	
	// Recalculate all positions with new config
	go s.RecalculatePositions(ctx)
	
	return nil
}

// LogStaffAction logs staff action
func (s *QueueService) LogStaffAction(ctx context.Context, entryID, staffID, staffName, action string, oldStatus, newStatus, oldPriority, newPriority, reason *string) error {
	log := &models.StaffQueueActionLog{
		ID:           utils.GenerateUUID(),
		QueueEntryID: entryID,
		StaffID:      staffID,
		StaffName:    &staffName,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		OldPriority:  oldPriority,
		NewPriority:  newPriority,
		Reason:       reason,
		Timestamp:    time.Now().UTC(),
	}

	return s.db.Create(log).Error
}

// RecordPositionHistory records position change
func (s *QueueService) RecordPositionHistory(ctx context.Context, entryID string, oldPos, newPos int, oldStatus, newStatus string, reason *string) error {
	history := &models.QueuePositionHistory{
		ID:           utils.GenerateUUID(),
		QueueEntryID: entryID,
		OldPosition:  oldPos,
		NewPosition:  newPos,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Reason:       reason,
		Timestamp:    time.Now().UTC(),
	}

	return s.db.Create(history).Error
}

// GetStaffActionLogs gets staff action logs
func (s *QueueService) GetStaffActionLogs(ctx context.Context, entryID string) ([]models.StaffQueueActionLog, error) {
	var logs []models.StaffQueueActionLog
	if err := s.db.Where("queue_entry_id = ?", entryID).
		Order("timestamp DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetQueueStatistics gets queue statistics
func (s *QueueService) GetQueueStatistics(ctx context.Context, date *time.Time) (*models.QueueStatsResponse, error) {
	targetDate := time.Now().UTC().Truncate(24 * time.Hour)
	if date != nil {
		targetDate = date.Truncate(24 * time.Hour)
	}

	var stats models.QueueStatistics
	if err := s.db.Where("date = ?", targetDate).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return empty stats
			return &models.QueueStatsResponse{
				Date: targetDate.Format("2006-01-02"),
			}, nil
		}
		return nil, err
	}

	return &models.QueueStatsResponse{
		Date:                 stats.Date.Format("2006-01-02"),
		TotalInQueue:         stats.TotalInQueue,
		WaitingCount:         stats.WaitingCount,
		InProgressCount:      stats.InProgressCount,
		ReadyCount:           stats.ReadyCount,
		CompletedToday:       stats.CompletedToday,
		CancelledToday:       stats.CancelledToday,
		AvgWaitTime:          stats.AvgWaitTime,
		AvgPreparationTime:   stats.AvgPreparationTime,
		CurrentLoad:          stats.CurrentLoad,
		OnTimeCompletionRate: stats.OnTimeCompletionRate,
	}, nil
}

// UpdateStatistics updates daily statistics
func (s *QueueService) UpdateStatistics(ctx context.Context) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	var stats models.QueueStatistics
	result := s.db.Where("date = ?", today).First(&stats)

	if result.Error != nil {
		stats = models.QueueStatistics{
			ID:   utils.GenerateUUID(),
			Date: today,
		}
	}

	// Count by status
	s.db.Model(&models.QueueEntry{}).Where("status = ? AND DATE(created_at) = ?", "WAITING", today).Count(&[]int64{int64(stats.WaitingCount)}[0])
	s.db.Model(&models.QueueEntry{}).Where("status = ? AND DATE(created_at) = ?", "IN_PROGRESS", today).Count(&[]int64{int64(stats.InProgressCount)}[0])
	s.db.Model(&models.QueueEntry{}).Where("status = ? AND DATE(created_at) = ?", "READY", today).Count(&[]int64{int64(stats.ReadyCount)}[0])
	s.db.Model(&models.QueueEntry{}).Where("status = ? AND DATE(created_at) = ?", "COMPLETED", today).Count(&[]int64{int64(stats.CompletedToday)}[0])
	s.db.Model(&models.QueueEntry{}).Where("status = ? AND DATE(created_at) = ?", "CANCELLED", today).Count(&[]int64{int64(stats.CancelledToday)}[0])

	stats.TotalInQueue = stats.WaitingCount + stats.InProgressCount + stats.ReadyCount
	stats.UpdatedAt = time.Now().UTC()

	if result.Error != nil {
		return s.db.Create(&stats).Error
	}
	return s.db.Save(&stats).Error
}

// GetUserQueueEntries gets all queue entries for a user
func (s *QueueService) GetUserQueueEntries(ctx context.Context, userID string) ([]models.QueueEntry, error) {
	var entries []models.QueueEntry
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// GetActiveQueueEntries gets all active entries
func (s *QueueService) GetActiveQueueEntries(ctx context.Context) ([]models.QueueEntry, error) {
	var entries []models.QueueEntry
	if err := s.db.Where("status IN ?", []string{"WAITING", "IN_PROGRESS", "READY"}).
		Order("position ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
