package handlers

import "net/http"

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
