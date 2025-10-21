package main

import (
	"TaskManager/internal/config"
	"TaskManager/internal/database"
	"TaskManager/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к БД
	db, err := database.ConnectDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка соединения с БД
	if err := database.PingDatabase(db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Успешное подключение к БД")

	// Создаем JWT сервис
	jwtService := services.NewJWTService(cfg.JWTSecret)

	// Создаем экземпляр приложения
	app := NewApp(db, cfg, jwtService)

	// Обслуживание статических файлов
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API маршруты
	http.HandleFunc("/api/register", app.apiMiddleware(app.registerHandler))
	http.HandleFunc("/api/login", app.apiMiddleware(app.loginHandler))
	http.HandleFunc("/api/refresh", app.apiMiddleware(app.refreshTokenHandler))
	http.HandleFunc("/api/health", app.apiMiddleware(app.healthHandler))

	// Защищенные API маршруты
	http.HandleFunc("/api/me", app.protectedApiMiddleware(app.meHandler))

	// Обработчик для задач - объединяем GET и POST в один маршрут
	http.HandleFunc("/api/tasks", app.protectedApiMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.getTasksHandler(w, r)
		case http.MethodPost:
			app.createTaskHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		}
	}))

	// Обработчик для операций с конкретной задачей (по ID)
	http.HandleFunc("/api/tasks/", app.protectedApiMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем task ID из URL
		path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
			return
		}

		//taskID := parts[0]

		// Обрабатываем разные методы
		switch r.Method {
		case http.MethodPut:
			// Проверяем, это toggle или обычное обновление
			if len(parts) > 1 && parts[1] == "toggle" {
				app.toggleTaskStatusHandler(w, r)
			} else {
				// Здесь можно добавить обработчик для обычного обновления задачи
				w.WriteHeader(http.StatusNotImplemented)
				json.NewEncoder(w).Encode(map[string]string{"error": "Обновление задачи пока не реализовано"})
			}
		case http.MethodDelete:
			app.deleteTaskHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		}
	}))

	// Страницы
	http.HandleFunc("/", app.registerFormHandler)
	http.HandleFunc("/dashboard", app.dashboardHandler)
	//http.HandleFunc("/dashboard", app.protectedApiMiddleware(app.dashboardHandler))

	port := cfg.GetServerPortString()
	server := &http.Server{
		Addr:    port,
		Handler: nil,
	}

	// Запуск сервера в горутине
	go func() {
		fmt.Printf("Сервер запущен на порту %s\n", port)
		fmt.Printf("Откройте в браузере: http://localhost%s\n", port)
		fmt.Printf("Dashboard: http://localhost%s/dashboard\n", port)
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
