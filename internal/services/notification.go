package services

import (
	"TaskManager/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type NotificationService struct {
	baseURL string
	client  *http.Client
}

func NewNotificationService(url string) *NotificationService {
	return &NotificationService{
		baseURL: url,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type NotificationRequest struct {
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	DueDate   time.Time `json:"due_date"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

func (ns *NotificationService) SendNotification(task *models.Task) error {
	notification := NotificationRequest{
		TaskID:    task.ID,
		UserID:    task.UserID,
		Title:     task.Title,
		Message:   task.Description,
		DueDate:   parseDueDate(task.DueDate),
		Priority:  task.Priority,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	resp, err := ns.client.Post(ns.baseURL+"/notifications", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	return nil
}

func parseDueDate(dueDate *string) time.Time {
	if dueDate == nil {
		return time.Now()
	}

	parsed, err := time.Parse("2006-01-02", *dueDate)
	if err != nil {
		return time.Now()
	}
	return parsed
}
