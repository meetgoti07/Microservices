package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gin-quickstart/config"
	"gin-quickstart/database"
	"gin-quickstart/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

func setupTestRouter() {
	gin.SetMode(gin.TestMode)
	router = gin.Default()
	routes.SetupRoutes(router)
}

func setupTestDB() {
	cfg := config.Load()
	database.InitDB(cfg)
	database.InitRedis(cfg)
}

func TestHealthCheck(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "queue-service", response["service"])
}

func TestGetCurrentQueue(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/queue/current", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetQueueStats(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/queue/stats", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetQueueStatsWithDate(t *testing.T) {
	setupTestRouter()

	date := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("/api/queue/stats?date=%s", date)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreateQueueEntry(t *testing.T) {
	setupTestRouter()

	payload := map[string]interface{}{
		"order_id":   "test-order-id",
		"user_id":    "test-user-id",
		"user_name":  "Test User",
		"user_phone": "+1234567890",
		"item_count": 3,
	}

	jsonData, _ := json.Marshal(payload)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/queue", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	router.ServeHTTP(w, req)

	// Will be 401 without proper auth, but tests the endpoint
	assert.True(t, w.Code == 401 || w.Code == 201)
}

func TestGetQueuePositionNotFound(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/queue/position/INVALID", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestGetQueueTokenNotFound(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/queue/token/INVALID", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestAdvanceQueueUnauthorized(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/queue/advance", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestUpdateQueueStatusUnauthorized(t *testing.T) {
	setupTestRouter()

	payload := map[string]interface{}{
		"status": "IN_PROGRESS",
	}

	jsonData, _ := json.Marshal(payload)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/queue/test-id/status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestCORS(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/queue/current", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
}

func TestInvalidDateFormat(t *testing.T) {
	setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/queue/stats?date=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}
