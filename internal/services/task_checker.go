package services

import (
	"TaskManager/internal/models"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type TaskChecker struct {
	db                  *sql.DB
	notificationService *NotificationService
	interval            time.Duration
}

func NewTaskChecker(db *sql.DB, notificationService *NotificationService, interval time.Duration) *TaskChecker {
	return &TaskChecker{
		db:                  db,
		notificationService: notificationService,
		interval:            interval,
	}
}

func (tc *TaskChecker) Start() {
	ticker := time.NewTicker(tc.interval)
	defer ticker.Stop()

	for range ticker.C {
		tc.checkDueTasks()
	}
}

func (tc *TaskChecker) checkDueTasks() {
	log.Println("Checking for due tasks...")

	query := `
        SELECT id, user_id, title, description, priority, due_date, created_at
        FROM tasks 
        WHERE deleted = false 
          AND notified = false 
          AND due_date IS NOT NULL 
          AND due_date <= CURRENT_DATE
          AND due_date >= CURRENT_DATE - INTERVAL '1 day'
    `

	rows, err := tc.db.Query(query)
	if err != nil {
		log.Printf("Error querying due tasks: %v", err)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.DueDate,
			&task.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	for _, task := range tasks {
		if err := tc.processTask(&task); err != nil {
			log.Printf("Error processing task %s: %v", task.ID, err)
		}
	}

	log.Printf("Processed %d due tasks", len(tasks))
}

func (tc *TaskChecker) processTask(task *models.Task) error {
	// Отправляем уведомление
	if err := tc.notificationService.SendNotification(task); err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	// Обновляем задачу как уведомленную
	_, err := tc.db.Exec(`
        UPDATE tasks 
        SET notified = true, notification_sent_at = $1 
        WHERE id = $2`,
		time.Now(), task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task notification status: %v", err)
	}

	log.Printf("Notification sent for task: %s", task.Title)
	return nil
}
