package services

import (
	"TaskManager/internal/models"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type TaskChecker struct {
	db            *sql.DB
	kafkaProducer *KafkaProducer
	interval      time.Duration
}

func NewTaskChecker(db *sql.DB, kafkaProducer *KafkaProducer, interval time.Duration) *TaskChecker {
	return &TaskChecker{
		db:            db,
		kafkaProducer: kafkaProducer,
		interval:      interval,
	}
}

func (tc *TaskChecker) Start() {
	log.Println("TaskChecker запущен")

	ticker := time.NewTicker(tc.interval)
	defer ticker.Stop()

	for range ticker.C {
		tc.checkDueTasks()
	}
}

func (tc *TaskChecker) checkDueTasks() {
	log.Println("Проверка запланированных задач...")

	query := `
        SELECT t.id, t.user_id, u.email, t.title, t.description, t.priority, t.due_date
        FROM tasks t
		INNER JOIN users u ON u.id = t.user_id
        WHERE t.deleted = false 
          AND t.notified = false 
          AND t.due_date IS NOT NULL 
          AND t.due_date <= CURRENT_DATE
          AND t.due_date >= CURRENT_DATE - INTERVAL '1 day'
		  AND email IS NOT NULL
    `

	rows, err := tc.db.Query(query)
	if err != nil {
		log.Printf("Ошибка поиска запланированных задач: %v", err)
		return
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.Task_id,
			&notification.User_id,
			&notification.Email,
			&notification.Title,
			&notification.Message,
			&notification.Priority,
			&notification.Due_date,
		)
		if err != nil {
			log.Printf("Ошибка сканирования задачи %s: %v", notification.ID, err)
			continue
		}
		notifications = append(notifications, notification)
	}

	for _, notification := range notifications {
		if err := tc.processTask(&notification); err != nil {
			log.Printf("Ошибка обработки задачи %s: %v", notification.Task_id, err)
		}
	}

	//log.Printf("Обработано %d запланированных задач", len(tasks))
}

func (tc *TaskChecker) processTask(notification *models.Notification) error {
	// Отправляем уведомление
	if err := tc.kafkaProducer.SendNotification(notification); err != nil {
		return fmt.Errorf("ошибка отпраки уведомления: %v", err)
	}

	// Обновляем задачу как уведомленную
	_, err := tc.db.Exec(`
        UPDATE tasks 
        SET notified = true, notification_sent_at = $1 
        WHERE id = $2`,
		time.Now(), notification.Task_id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления времени отправки уведомления у задачи: %v", err)
	}

	log.Printf("Уведомление отправлено для задачи: %s", notification.Task_id)
	return nil
}
