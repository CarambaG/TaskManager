package handlers

import (
	"TaskManager/internal/config"
	"TaskManager/internal/models"
	"TaskManager/internal/services"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type App struct {
	db         *sql.DB
	cfg        *config.Config
	jwtService *services.JWTService
}

func NewApp(db *sql.DB, cfg *config.Config, jwtService *services.JWTService) *App {
	return &App{
		db:         db,
		cfg:        cfg,
		jwtService: jwtService,
	}
}

// Обработчик обновления токена
func (a *App) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	newToken, err := a.jwtService.RefreshToken(req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидный токен"})
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: newToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик для получения информации о текущем пользователе
func (a *App) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Получаем пользователя из контекста (из authMiddleware)
	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Получаем полную информацию о пользователе из БД
	var dbUser models.User
	err := a.db.QueryRow("SELECT id, login, create_at FROM users WHERE id = $1", userClaims.UserID).Scan(
		&dbUser.ID, &dbUser.Login, &dbUser.CreateAt,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения данных пользователя"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbUser)
}

// Обработчик получения задач
func (a *App) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Здесь будет логика получения задач из БД
	// Пока возвращаем пустой массив
	tasks := []Task{}

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

	var taskData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		DueDate     string `json:"due_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	// Валидация данных
	if taskData.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Название задачи не может быть пустым"})
		return
	}

	// Здесь будет логика создания задачи в БД
	// Пока возвращаем успешный ответ

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

	taskID := parts[0]

	userClaims, ok := r.Context().Value("user").(*services.Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Невалидные данные пользователя"})
		return
	}

	// Здесь будет логика переключения статуса задачи в БД

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

	// Здесь будет логика удаления задачи из БД

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
