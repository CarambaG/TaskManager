package handlers

import (
	"TaskManager/internal/controllers"
	"TaskManager/internal/models"
	"TaskManager/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	tasks, err := controllers.GetTasks(&userClaims.UserID, a.db)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при поиске задач пользователя"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
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

	/*var taskData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		DueDate     string `json:"due_date"`
	}*/

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
