package handlers

import (
	"TaskManager/internal/config"
	"TaskManager/internal/models"
	"TaskManager/internal/services"
	"database/sql"
	"encoding/json"
	"net/http"
)

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
	err := a.db.QueryRow("SELECT id, login, email, create_at FROM users WHERE id = $1", userClaims.UserID).Scan(
		&dbUser.ID, &dbUser.Login, &dbUser.Email, &dbUser.CreateAt,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения данных пользователя"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbUser)
}
