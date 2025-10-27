package models

import "time"

type Task struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Status             string    `json:"status"`
	Priority           string    `json:"priority"`
	DueDate            *string   `json:"due_date"`
	Notified           bool      `json:"notified"`
	NotificationSentAt time.Time `json:"notification_sent_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
