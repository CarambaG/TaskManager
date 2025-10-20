package main

import (
	"TaskManager/internal/config"
	"TaskManager/internal/controllers"
	"TaskManager/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type App struct {
	db *sql.DB
}

func main() {
	cfg := config.Load()
	var db *sql.DB = nil

	//Подключение к БД
	db, err := database.ConnectDatabase(cfg)
	if err != nil {
		//fmt.Println(cfg)
		log.Fatal(err)
	}
	defer db.Close()

	//Соединение с БД
	if err := database.PingDatabase(db); err != nil {
		//fmt.Println(cfg)
		log.Fatal(err)
	}
	fmt.Println("Успешное подключение к БД")

	// Создаем экземпляр приложения
	app := &App{db: db}

	// Проверяем существование папки static
	/*if _, err := os.Stat("./static"); os.IsNotExist(err) {
		log.Fatal("Папка 'static' не найдена.")
	}*/

	// Обслуживание статических файлов
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API маршруты
	http.HandleFunc("/api/register", loggingMiddleware(enableCORS(app.registerHandler)))
	http.HandleFunc("/api/login", loggingMiddleware(enableCORS(app.loginHandler)))
	http.HandleFunc("/api/health", loggingMiddleware(enableCORS(app.healthHandler)))

	// Главная страница
	http.HandleFunc("/", app.indexHandler)

	port := cfg.GetServerPortString()
	server := &http.Server{
		Addr:    port,
		Handler: nil,
	}

	go func() {
		fmt.Printf("Сервер запущен на порту %s\n", port)
		fmt.Printf("Откройте в браузере: http://localhost%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nЗавершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении сервера: %v", err)
	}
	fmt.Println("Сервер корректно завершен")
}

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Middleware для логирования
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(lrw, r)

		duration := time.Since(start)
		log.Printf("Completed %s %s %d %v", r.Method, r.URL.Path, lrw.statusCode, duration)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Обработчик главной страницы
func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Отдаем HTML файл
	http.ServeFile(w, r, "./internal/views/registerForm/index.html")
}

// Обработчик регистрации
func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
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
func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
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

	// Формируем ответ
	response := struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	}{
		ID:    authUser.ID,
		Login: authUser.Login,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Простой health check
func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	response := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
