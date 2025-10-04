package handlers

import (
	"net/http"
	"time"

	"gin-quickstart/models"
	"gin-quickstart/services"

	"github.com/gin-gonic/gin"
)

type QueueHandler struct {
	service *services.QueueService
}

func NewQueueHandler() *QueueHandler {
	return &QueueHandler{
		service: services.NewQueueService(),
	}
}

// GetUserFromContext extracts user from context (set by auth middleware)
func GetUserFromContext(c *gin.Context) (string, string, string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", "", "", false
	}
	
	userName, _ := c.Get("user_name")
	userRole, _ := c.Get("user_role")
	
	return userID.(string), userName.(string), userRole.(string), true
}

// CreateQueueEntry creates a new queue entry
// POST /api/queue
func (h *QueueHandler) CreateQueueEntry(c *gin.Context) {
	var req models.CreateQueueEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	entry, err := h.service.CreateQueueEntry(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to create queue entry",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Queue entry created successfully",
		Data:    entry,
	})
}

// GetQueuePosition gets position for a token
// GET /api/queue/position/:token
func (h *QueueHandler) GetQueuePosition(c *gin.Context) {
	token := c.Param("token")

	position, err := h.service.GetQueuePosition(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Queue entry not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, position)
}

// GetQueueEntryByToken gets queue entry by token
// GET /api/queue/token/:token
func (h *QueueHandler) GetQueueEntryByToken(c *gin.Context) {
	token := c.Param("token")

	entry, err := h.service.GetQueueEntryByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Queue entry not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetQueueEntryByOrderID gets queue entry by order ID
// GET /api/queue/order/:orderId
func (h *QueueHandler) GetQueueEntryByOrderID(c *gin.Context) {
	orderID := c.Param("orderId")

	entry, err := h.service.GetQueueEntryByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Queue entry not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// GetCurrentQueue gets current queue state
// GET /api/queue/current
func (h *QueueHandler) GetCurrentQueue(c *gin.Context) {
	queue, err := h.service.GetCurrentQueue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get current queue",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// UpdateQueueStatus updates queue entry status (Staff only)
// PUT /api/queue/:id/status
func (h *QueueHandler) UpdateQueueStatus(c *gin.Context) {
	entryID := c.Param("id")
	userID, userName, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req models.UpdateQueueStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateQueueStatus(c.Request.Context(), entryID, &req, userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update queue status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Queue status updated successfully",
	})
}

// UpdateQueuePriority updates queue entry priority (Staff only)
// PUT /api/queue/:id/priority
func (h *QueueHandler) UpdateQueuePriority(c *gin.Context) {
	entryID := c.Param("id")
	userID, userName, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req models.UpdateQueuePriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateQueuePriority(c.Request.Context(), entryID, &req, userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update queue priority",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Queue priority updated successfully",
	})
}

// AssignStaff assigns staff to queue entry (Staff only)
// POST /api/queue/:id/assign
func (h *QueueHandler) AssignStaff(c *gin.Context) {
	entryID := c.Param("id")
	userID, userName, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req models.AssignStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.AssignStaff(c.Request.Context(), entryID, &req, userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to assign staff",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Staff assigned successfully",
	})
}

// AdvanceQueue advances the queue (Staff only)
// POST /api/queue/advance
func (h *QueueHandler) AdvanceQueue(c *gin.Context) {
	userID, userName, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.service.AdvanceQueue(c.Request.Context(), userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to advance queue",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Queue advanced successfully",
	})
}

// GetQueueStatistics gets queue statistics
// GET /api/queue/stats
func (h *QueueHandler) GetQueueStatistics(c *gin.Context) {
	var date *time.Time
	if dateStr := c.Query("date"); dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Invalid date format",
				Message: "Use YYYY-MM-DD format",
			})
			return
		}
		date = &parsedDate
	}

	stats, err := h.service.GetQueueStatistics(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get statistics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserQueueEntries gets all queue entries for the authenticated user
// GET /api/queue/user/me
func (h *QueueHandler) GetUserQueueEntries(c *gin.Context) {
	userID, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	entries, err := h.service.GetUserQueueEntries(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get user queue entries",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// GetActiveQueueEntries gets all active queue entries (Public for admin)
// GET /api/queue
func (h *QueueHandler) GetActiveQueueEntries(c *gin.Context) {
	entries, err := h.service.GetActiveQueueEntries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get active queue entries",
			Message: err.Error(),
		})
		return
	}

	// Return in paginated format expected by frontend
	response := map[string]interface{}{
		"entries":         entries,
		"total":           len(entries),
		"page":            1,
		"pageSize":        len(entries),
		"totalPages":      1,
		"hasNextPage":     false,
		"hasPreviousPage": false,
	}

	c.JSON(http.StatusOK, response)
}

// GetStaffActionLogs gets staff action logs for an entry (Staff only)
// GET /api/queue/:id/logs
func (h *QueueHandler) GetStaffActionLogs(c *gin.Context) {
	entryID := c.Param("id")

	logs, err := h.service.GetStaffActionLogs(c.Request.Context(), entryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get action logs",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetConfiguration gets queue configuration (Staff only)
// GET /api/queue/config
func (h *QueueHandler) GetConfiguration(c *gin.Context) {
	config, err := h.service.GetConfiguration(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get configuration",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfiguration updates queue configuration (Admin only)
// PUT /api/queue/config
func (h *QueueHandler) UpdateConfiguration(c *gin.Context) {
	userID, _, _, ok := GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var config models.QueueConfiguration
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateConfiguration(c.Request.Context(), &config, userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to update configuration",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Configuration updated successfully",
		Data:    config,
	})
}

// RecalculatePositions recalculates all positions (Staff only)
// POST /api/queue/recalculate
func (h *QueueHandler) RecalculatePositions(c *gin.Context) {
	if err := h.service.RecalculatePositions(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to recalculate positions",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Positions recalculated successfully",
	})
}
