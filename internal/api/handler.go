package api

import (
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cDNS/internal/config"
	ldns "cDNS/internal/dns"
	"cDNS/internal/task"
)

var (
	startTime = time.Now()
	version   = "dev"
)

var taskManager = task.Manager

type BackgroundTask = task.BackgroundTask

type Handler struct {
	logger *zap.Logger
	router *gin.Engine
}

func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{
		logger: logger,
		router: gin.New(),
	}
}

func (h *Handler) SetupRoutes() {
	h.router.Use(h.ginLogger())
	h.router.Use(gin.Recovery())
	v1 := h.router.Group("/api/v1")
	{
		v1.GET("/health", h.HealthCheck)
		v1.GET("/dns-servers", h.GetDNSServers)
		v1.POST("/query", h.QueryEndpoint)
		v1.POST("/query/background", h.BackgroundQueryEndpoint)
		v1.GET("/task/:id", h.GetTaskEndpoint)
		v1.GET("/tasks", h.GetTasksEndpoint)
	}
}

func (h *Handler) GetRouter() *gin.Engine {
	return h.router
}

func (h *Handler) ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		h.logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
		)
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": version,
		"uptime":  time.Since(startTime).String(),
	})
}

func (h *Handler) GetDNSServers(c *gin.Context) {
	servers := make([]map[string]string, 0, len(ldns.PopularDNSServers))
	dnsInfo := map[string]string{
		"8.8.8.8":         "Google DNS",
		"8.8.4.4":         "Google DNS",
		"1.1.1.1":         "Cloudflare DNS",
		"1.0.0.1":         "Cloudflare DNS",
		"9.9.9.9":         "Quad9 DNS",
		"149.112.112.112": "Quad9 DNS",
		"208.67.222.222":  "OpenDNS",
		"208.67.220.220":  "OpenDNS",
		"76.76.19.19":     "Alternate DNS",
		"76.223.100.101":  "Alternate DNS",
		"94.140.14.14":    "AdGuard DNS",
		"94.140.15.15":    "AdGuard DNS",
		"77.88.8.8":       "Yandex DNS",
		"77.88.8.1":       "Yandex DNS",
	}
	for _, server := range ldns.PopularDNSServers {
		servers = append(servers, map[string]string{
			"ip":   server,
			"name": dnsInfo[server],
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"dns_servers": servers,
	})
}

func (h *Handler) QueryEndpoint(c *gin.Context) {
	var req struct {
		Domain      string   `json:"domain" binding:"required"`
		Nameservers []string `json:"nameservers" binding:"required"`
		Timeout     int      `json:"timeout,omitempty"`
		Retries     int      `json:"retries,omitempty"`
		Filter      []string `json:"filter,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg := config.Config{
		Timeout:      time.Duration(req.Timeout) * time.Second,
		Retries:      req.Retries,
		RecordFilter: req.Filter,
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}

	// Ensure domain is fully qualified
	domain := dns.Fqdn(req.Domain)

	if !ldns.IsValidDomain(domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain"})
		return
	}
	nameservers := ldns.PrepareNameservers(req.Nameservers)
	if len(nameservers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid nameservers provided"})
		return
	}
	var results []ldns.Result
	for _, ns := range nameservers {
		result := ldns.Nameserver(domain, ns, cfg)
		results = append(results, result)
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *Handler) BackgroundQueryEndpoint(c *gin.Context) {
	var req struct {
		Domain      string   `json:"domain" binding:"required"`
		Nameservers []string `json:"nameservers" binding:"required"`
		Timeout     int      `json:"timeout,omitempty"`
		Retries     int      `json:"retries,omitempty"`
		Filter      []string `json:"filter,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Create task
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	taskObj := &task.BackgroundTask{
		ID:          taskID,
		Domain:      req.Domain,
		Nameservers: req.Nameservers,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	taskManager.AddTask(taskObj)
	go task.ProcessBackgroundTask(taskID, req)
	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": "pending"})
}

func (h *Handler) saveResultsToFile(results []ldns.Result, filename string) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (h *Handler) GetTaskEndpoint(c *gin.Context) {
	taskID := c.Param("id")
	taskObj, exists := taskManager.GetTask(taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, taskObj)
}

func (h *Handler) GetTasksEndpoint(c *gin.Context) {
	status := c.Query("status")
	tasks := taskManager.GetTasks(status)
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
