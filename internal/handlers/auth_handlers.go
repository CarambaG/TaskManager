package handlers

import (
	"TaskManager/internal/controllers"
	"encoding/json"
	"net/http"
)

// Обработчик главной страницы
func (a *App) RegisterFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Отдаем HTML файл
	http.ServeFile(w, r, "./internal/views/registerForm/index.html")
}

// Обработчик регистрации
func (a *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный JSON"})
		return
	}

	// Создаем пользователя
	newUser, err := controllers.CreateUser(a.db, req.Login, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Формируем ответ
	response := struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	}{
		ID:    newUser.ID,
		Login: newUser.Login,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик аутентификации
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	// Аутентифицируем пользователя
	authUser, err := controllers.Authenticate(a.db, req.Login, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
		return
	}

	// Генерируем JWT токен
	token, err := a.jwtService.GenerateToken(authUser)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка генерации токена"})
		return
	}

	// Формируем ответ
	response := struct {
		//ID    string `json:"id"`
		Token string `json:"token"`
	}{
		//ID:    authUser.ID,
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
