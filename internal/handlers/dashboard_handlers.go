package handlers

import (
	"TaskManager/internal/controllers"
	"TaskManager/internal/models"
	"TaskManager/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Обработчик dashboard
func (a *App) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем аутентификацию через middleware
	// Если пользователь не аутентифицирован, middleware уже вернет ошибку

	// Отдаем страницу dashboard
	http.ServeFile(w, r, "./internal/views/dashboard/index.html")
}

// Обработчик получения задач
func (a *App) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Получение задач из БД
	tasks, err := controllers.GetTasksDataBase(&userClaims.UserID, a.db)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задач пользователя"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Обработчик получения задачи
func (a *App) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем taskId из url
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Получение задач из БД
	taskData, err := controllers.GetTaskDataBase(a.db, &userClaims.UserID, &parts[0])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задачи пользователя"})
		return
	}

	response := struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Priority    string  `json:"priority"`
		DueDate     *string `json:"due_date"`
	}{
		Title:       taskData.Title,
		Description: taskData.Description,
		Priority:    taskData.Priority,
		DueDate:     taskData.DueDate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик создания задачи
func (a *App) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	newTaskData := models.Task{}

	if err := json.NewDecoder(r.Body).Decode(&newTaskData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	// Валидация данных
	if newTaskData.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Название задачи не может быть пустым"})
		return
	}

	newTaskData.UserID = userClaims.UserID
	err := controllers.CreateTaskDataBase(a.db, &newTaskData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при создании задачи"})
		return
	}

	response := struct {
		Message string `json:"message"`
		TaskID  string `json:"task_id"`
		UserID  string `json:"user_id"`
	}{
		Message: "Задача успешно создана",
		TaskID:  "temp-id", // Заменится на реальный ID из БД
		UserID:  userClaims.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Обработчик переключения статуса задачи
func (a *App) ToggleTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем task ID из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
		return
	}

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	taskID := parts[0]

	// Изменение задачи в БД
	err := controllers.ToggleTaskStatusDataBase(a.db, &taskID, &userClaims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := struct {
		Message string `json:"message"`
		TaskID  string `json:"task_id"`
		UserID  string `json:"user_id"`
	}{
		Message: "Статус задачи обновлен",
		TaskID:  taskID,
		UserID:  userClaims.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик удаления задачи
func (a *App) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем task ID из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
		return
	}

	taskID := parts[0]

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Удаление задачи в БД
	err := controllers.DeleteTaskDataBase(a.db, &taskID, &userClaims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := struct {
		Message string `json:"message"`
		TaskID  string `json:"task_id"`
		UserID  string `json:"user_id"`
	}{
		Message: "Задача удалена",
		TaskID:  taskID,
		UserID:  userClaims.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) SaveTaskDataHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
		return
	}

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	newTaskData := models.Task{}
	if err := json.NewDecoder(r.Body).Decode(&newTaskData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	err := controllers.SavaTaskDB(a.db, &userClaims.UserID, &parts[0], &newTaskData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := struct {
		Message string `json:"message"`
		TaskID  string `json:"task_id"`
		UserID  string `json:"user_id"`
	}{
		Message: "Задача изменена",
		TaskID:  parts[0],
		UserID:  userClaims.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) SaveUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user").(*services.Claims).UserID
	body := struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	// Валидация нового пароля
	if body.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Новый пароль не может быть пустым"})
		return
	}

	if len(body.NewPassword) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пароль должен содержать минимум 6 символов"})
		return
	}

	// Начало смены пароля
	tx, err := a.db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене пароля"})
		return
	}
	defer tx.Rollback()

	query1 := `
		SELECT pass FROM users WHERE id = $1
	`

	passDB := ""
	tx.QueryRow(query1, user_id).Scan(&passDB)

	if !controllers.CheckPasswordHash(body.CurrentPassword, passDB) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный пароль"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 12)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене пароля"})
	}

	query2 := `
		UPDATE users SET pass = $1 WHERE id = $2
	`

	_, err = tx.Exec(query2, hash, user_id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене пароля"})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене пароля"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "пароль изменен"})
}

func (a *App) SaveUserEmailHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user").(*services.Claims).UserID
	body := struct {
		Email string `json:"new_email"`
		Pass  string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	if body.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "почта не должна быть пустой"})
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене почты"})
		return
	}
	defer tx.Rollback()

	query1 := `
		SELECT pass FROM users WHERE id = $1
	`

	// Проверка пароля
	passDB := ""
	err = tx.QueryRow(query1, user_id).Scan(&passDB)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене почты"})
		return
	}

	if !controllers.CheckPasswordHash(body.Pass, passDB) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "неверный пароль"})
		return
	}

	query2 := `
		UPDATE users SET email = $1 WHERE id = $2
	`

	_, err = tx.Exec(query2, body.Email, user_id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене почты"})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ошибка при смене почты"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "почта изменена"})
}
