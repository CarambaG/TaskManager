package controllers

import (
	"TaskManager/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func GetTasks(UserID *string, db *sql.DB) (tasks []models.Task, err error) {
	tasks = []models.Task{}
	query := `
        SELECT 
            id,
    		title,
    		description,
    		status,
    		priority,
    		due_date,
    		created_at,
    		updated_at
        FROM tasks 
        WHERE deleted = false
        	and user_id = $1
    `
	rows, err := db.Query(query, *UserID)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("ошибка запроса к БД: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func ToggleTaskStatusDataBase(db *sql.DB, taskID *string, UserID *string) (err error) {
	var (
		task_ID     string
		task_Status string
	)
	query1 := `
		SELECT 
			id,
			status
		FROM tasks
		WHERE
			id = $1
			and user_id = $2
	`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем наличие задачи у конкретного пользователя
	err = tx.QueryRow(query1, *taskID, *UserID).Scan(&task_ID, &task_Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("задача не найдена")
		}
		return err
	}

	// Определяем новый статус
	var newStatus string
	switch task_Status {
	case "active":
		newStatus = "completed"
	case "completed":
		newStatus = "active"
	default:
		newStatus = "active"
	}

	query2 := `
		UPDATE tasks
		SET status = $1,
		    updated_at = now()
		WHERE id = $2
	`

	// Обновляем статус задачи
	_, err = tx.Exec(query2, newStatus, *taskID)
	if err != nil {
		return err
	}

	// Подтверждаем транзакцию
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DeleteTaskDataBase(db *sql.DB, taskID *string, UserID *string) (err error) {
	query2 := `
		UPDATE tasks
		SET deleted = true
		WHERE id = $1
			AND user_id = $2
	`

	// Проставляем флаг удаления задачи
	result, err := db.Exec(query2, *taskID, *UserID)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("задача не найдена")
	}

	return nil
}

func CreateTaskDataBase(db *sql.DB, taskData *models.Task) (err error) {
	query := `
		INSERT INTO tasks (user_id, deleted, title, description, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	fmt.Println(false, '\n', taskData.Title, '\n', taskData.Description, '\n', taskData.Status, '\n', taskData.Priority, '\n', taskData.DueDate, '\n', time.Now(), '\n', time.Now())

	// Вставляем новую задачу в БД
	_, err = db.Exec(query,
		taskData.UserID,
		false,
		taskData.Title,
		taskData.Description,
		"active",
		taskData.Priority,
		taskData.DueDate,
		time.Now(),
		time.Now())
	if err != nil {
		return err
	}

	return nil
}
