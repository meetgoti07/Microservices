package routes

import (
	"gin-quickstart/handlers"
	"gin-quickstart/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	queueHandler := handlers.NewQueueHandler()

	// Apply CORS
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "queue-service",
		})
	})

	// Public routes
	public := router.Group("/api/queue")
	{
		// Get all active queue entries (public - for display)
		public.GET("", queueHandler.GetActiveQueueEntries)
		
		// Get queue position by token (public)
		public.GET("/position/:token", queueHandler.GetQueuePosition)
		
		// Get queue entry by token (public)
		public.GET("/token/:token", queueHandler.GetQueueEntryByToken)
		
		// Get current queue state (public - for display)
		public.GET("/current", queueHandler.GetCurrentQueue)
		
		// Get queue statistics (public - for display)
		public.GET("/stats", queueHandler.GetQueueStatistics)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/queue")
	protected.Use(middleware.AuthMiddleware())
	{
		// Create queue entry (authenticated users)
		protected.POST("", queueHandler.CreateQueueEntry)
		
		// Get queue entry by order ID
		protected.GET("/order/:orderId", queueHandler.GetQueueEntryByOrderID)
		
		// Get user's own queue entries
		protected.GET("/user/me", queueHandler.GetUserQueueEntries)
	}

	// Staff routes (require staff role)
	staff := router.Group("/api/queue")
	staff.Use(middleware.AuthMiddleware(), middleware.StaffOnlyMiddleware())
	{
		// Update queue status
		staff.PATCH("/:id/status", queueHandler.UpdateQueueStatus)
		
		// Update queue priority
		staff.PUT("/:id/priority", queueHandler.UpdateQueuePriority)
		
		// Assign staff to queue entry
		staff.POST("/:id/assign", queueHandler.AssignStaff)
		
		// Advance queue
		staff.POST("/advance", queueHandler.AdvanceQueue)
		
		// Get staff action logs
		staff.GET("/:id/logs", queueHandler.GetStaffActionLogs)
		
		// Get configuration
		staff.GET("/config", queueHandler.GetConfiguration)
		
		// Recalculate positions
		staff.POST("/recalculate", queueHandler.RecalculatePositions)
	}

	// Admin routes (require admin role)
	admin := router.Group("/api/queue")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminOnlyMiddleware())
	{
		// Update configuration
		admin.PUT("/config", queueHandler.UpdateConfiguration)
	}
}
