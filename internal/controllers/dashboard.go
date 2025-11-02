package controllers

import (
	"TaskManager/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func GetTasksDataBase(UserID *string, db *sql.DB) (tasks []models.Task, err error) {
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
        	AND user_id = $1
        ORDER BY created_at DESC
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

	// Проверяем наличие задачи у пользователя
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

	//Проверяем наличие задачи
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

func GetTaskDataBase(db *sql.DB, UserID *string, TaskId *string) (taskData models.Task, err error) {
	taskData = models.Task{}
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
        	and id = $2
    `

	// Получение задачи и БД
	err = db.QueryRow(query, *UserID, *TaskId).Scan(
		&taskData.ID,
		&taskData.Title,
		&taskData.Description,
		&taskData.Status,
		&taskData.Priority,
		&taskData.DueDate,
		&taskData.CreatedAt,
		&taskData.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return taskData, errors.New("задача не найдена")
		}
		return taskData, fmt.Errorf("ошибка запроса к БД: %v", err)
	}

	return taskData, nil
}

func SavaTaskDB(db *sql.DB, UserID *string, TaskID *string, newTaskData *models.Task) (err error) {
	query := `
		UPDATE tasks
		SET 
		    title = $1,
		    description = $2,
		    priority = $3,
		    due_date = $4
		WHERE deleted = false
		  	AND user_id = $5 
			AND id = $6
	`

	result, err := db.Exec(query,
		newTaskData.Title,
		newTaskData.Description,
		newTaskData.Priority,
		newTaskData.DueDate,
		*UserID,
		*TaskID,
	)
	fmt.Println(newTaskData.Title, "\n",
		newTaskData.Description, "\n",
		newTaskData.Priority, "\n",
		newTaskData.DueDate, "\n",
		*UserID, "\n",
		*TaskID, "\n", err)
	if err != nil {
		return err
	}

	// Проверяем наличие задачи
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("задача не найдена")
	}

	return nil
}
