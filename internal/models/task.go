package models

import (
	"errors"
	"regexp"
	"time"
)

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

func (t *Task) Validate() error {
	if err := t.validateId(); err != nil {
		return err
	}

	if err := t.validateUserId(); err != nil {
		return err
	}

	if err := t.validateTitle(); err != nil {
		return err
	}

	if err := t.validatePriority(); err != nil {
		return err
	}

	if err := t.validateDueDate(); err != nil {
		return err
	}

	if err := t.validateStatus(); err != nil {
		return err
	}

	return nil
}

// validateId проверяем Id задачи
func (t *Task) validateId() error {
	if t.ID == "" {
		return nil
	}

	matched, err := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, t.ID)
	if err != nil || !matched {
		return errors.New("Invalid ID")
	}

	return nil
}

// validateUserId проверяет UserId
func (t *Task) validateUserId() error {
	if t.UserID == "" {
		return nil
	}

	matched, err := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, t.UserID)
	if err != nil || !matched {
		return errors.New("Invalid UserID")
	}

	return nil
}

// validateTitle проверяет заголовок задачи
func (t *Task) validateTitle() error {
	if t.Title == "" {
		return errors.New("title cannot be empty")
	}

	if len(t.Title) > 200 {
		return errors.New("title cannot exceed 200 characters")
	}

	return nil
}

// validatePriority проверяет приоритет задачи
func (t *Task) validatePriority() error {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	if !validPriorities[t.Priority] {
		return errors.New("priority must be one of: low, medium, high")
	}

	return nil
}

// validateDueDate проверяет дату выполнения
func (t *Task) validateDueDate() error {
	if t.DueDate == nil {
		return nil // nil - допустимое значение (нет даты)
	}

	dateStr := *t.DueDate
	if dateStr == "" {
		return nil // пустая строка - тоже допустима
	}

	// Проверяем формат даты (YYYY-MM-DD)
	matched, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr)
	if err != nil || !matched {
		return errors.New("due_date must be in format YYYY-MM-DD")
	}

	// Парсим дату
	due, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("invalid due_date format")
	}

	// Проверяем, что дата не в прошлом
	today := time.Now().Truncate(24 * time.Hour)
	due = due.Truncate(24 * time.Hour)

	if due.Before(today) {
		return errors.New("due_date cannot be in the past")
	}

	return nil
}

// validateStatus проверяет статус задачи
func (t *Task) validateStatus() error {
	if t.Status == "" {
		return nil // пустой статус - допустимо (будет установлен по умолчанию)
	}

	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
	}

	if !validStatuses[t.Status] {
		return errors.New("status must be one of: active, completed")
	}

	return nil
}
