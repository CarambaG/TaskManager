package models

import "time"

type Notification struct {
	ID         string    `json:"id"`
	Task_id    string    `json:"task_id"`
	User_id    string    `json:"user_id"`
	Email      string    `json:"email"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	Status     string    `json:"status"`
	Due_date   *string   `json:"due_date"`
	Priority   string    `json:"priority"`
	Created_at time.Time `json:"created_at"`
	Sent_at    time.Time `json:"sent_at"`
}
