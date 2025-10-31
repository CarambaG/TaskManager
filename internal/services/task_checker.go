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
	log.Println("Проверка запланированных задач...")

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
		log.Printf("Ошибка поиска запланированных задач: %v", err)
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
			log.Printf("Ошибка сканирования задачи %s: %v", task.ID, err)
			continue
		}
		tasks = append(tasks, task)
	}

	for _, task := range tasks {
		if err := tc.processTask(&task); err != nil {
			log.Printf("Ошибка обработки задачи %s: %v", task.ID, err)
		}
	}

	//log.Printf("Обработано %d запланированных задач", len(tasks))
}

func (tc *TaskChecker) processTask(task *models.Task) error {
	// Отправляем уведомление
	if err := tc.notificationService.SendNotification(task); err != nil {
		return fmt.Errorf("ошибка отпраки уведомления: %v", err)
	}

	// Обновляем задачу как уведомленную
	_, err := tc.db.Exec(`
        UPDATE tasks 
        SET notified = true, notification_sent_at = $1 
        WHERE id = $2`,
		time.Now(), task.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления времени отправки уведомления у задачи: %v", err)
	}

	log.Printf("Уведомление отправлено для задачи: %s", task.ID)
	return nil
}
