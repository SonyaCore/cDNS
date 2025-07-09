package task

import (
	"cDNS/internal/config"
	"cDNS/internal/dns"
	"cDNS/internal/logger"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
	"strings"
	"sync"
	"time"
)

type BackgroundTask struct {
	ID          string       `json:"id"`
	Domain      string       `json:"domain"`
	Nameservers []string     `json:"nameservers"`
	Status      string       `json:"status"` // pending, running, completed, failed
	Results     []dns.Result `json:"results,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Error       string       `json:"error,omitempty"`
}

type TaskManager struct {
	tasks map[string]*BackgroundTask
	mutex sync.RWMutex
}

var Manager = &TaskManager{
	tasks: make(map[string]*BackgroundTask),
}

func (tm *TaskManager) AddTask(task *BackgroundTask) {
	tm.mutex.Lock()
	tm.tasks[task.ID] = task
	tm.mutex.Unlock()
}

func (tm *TaskManager) GetTask(id string) (*BackgroundTask, bool) {
	tm.mutex.RLock()
	task, exists := tm.tasks[id]
	tm.mutex.RUnlock()
	return task, exists
}

func (tm *TaskManager) GetTasks(status string) []*BackgroundTask {
	tm.mutex.RLock()
	tasks := make([]*BackgroundTask, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		if status == "" || task.Status == status {
			tasks = append(tasks, task)
		}
	}
	tm.mutex.RUnlock()
	return tasks
}

func ProcessBackgroundTask(taskID string, req struct {
	Domain      string   `json:"domain" binding:"required"`
	Nameservers []string `json:"nameservers" binding:"required"`
	Timeout     int      `json:"timeout,omitempty"`
	Retries     int      `json:"retries,omitempty"`
	Filter      []string `json:"filter,omitempty"`
}) {
	Manager.mutex.Lock()
	task := Manager.tasks[taskID]
	task.Status = "running"
	Manager.mutex.Unlock()
	defer func() {
		completedAt := time.Now()
		Manager.mutex.Lock()
		task.CompletedAt = &completedAt
		Manager.mutex.Unlock()
	}()
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
	domain := req.Domain
	if !dns.IsValidDomain(domain) {
		Manager.mutex.Lock()
		task.Status = "failed"
		task.Error = "Invalid domain"
		Manager.mutex.Unlock()
		return
	}
	nameservers := dns.PrepareNameservers(req.Nameservers)
	if len(nameservers) == 0 {
		Manager.mutex.Lock()
		task.Status = "failed"
		task.Error = "No valid nameservers provided"
		Manager.mutex.Unlock()
		return
	}
	var results []dns.Result
	for _, ns := range nameservers {
		result := dns.Nameserver(domain, ns, cfg)
		results = append(results, result)
	}
	filename := fmt.Sprintf("dns_results_%s_%d.json", strings.ReplaceAll(req.Domain, ".", "_"), time.Now().Unix())
	if err := saveResultsToFile(results, filename); err != nil {
		logger.GetLogger().Error("Failed to save results to file", zap.Error(err))
		Manager.mutex.Lock()
		task.Status = "failed"
		task.Error = "Failed to save results to file"
		Manager.mutex.Unlock()
		return
	}
	Manager.mutex.Lock()
	task.Status = "completed"
	task.Results = results
	Manager.mutex.Unlock()
	logger.GetLogger().Info("Background task completed", zap.String("task_id", taskID), zap.String("file", filename))
}

func saveResultsToFile(results []dns.Result, filename string) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}
