package main

import (
	"TaskManager/internal/config"
	"TaskManager/internal/database"
	"TaskManager/internal/handlers"
	"TaskManager/internal/services"
	"context"
	"encoding/json"
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
	log.Println("Успешное подключение к БД")

	// Создаем JWT сервис
	jwtService := services.NewJWTService(cfg.JWTSecret)

	// Создаем экземпляр приложения
	app := handlers.NewApp(db, cfg, jwtService)

	// Обслуживание статических файлов
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API маршруты
	http.HandleFunc("/api/register", app.ApiMiddleware(app.RegisterHandler))
	http.HandleFunc("/api/login", app.ApiMiddleware(app.LoginHandler))
	http.HandleFunc("/api/refresh", app.ApiMiddleware(app.RefreshTokenHandler))
	http.HandleFunc("/api/health", app.ApiMiddleware(app.HealthHandler))

	// Защищенные API маршруты
	http.HandleFunc("/api/me", app.ProtectedApiMiddleware(app.MeHandler))

	// Обработчик для задач - объединяем GET и POST в один маршрут
	http.HandleFunc("/api/tasks", app.ProtectedApiMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetTasksHandler(w, r)
		case http.MethodPost:
			app.CreateTaskHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		}
	}))

	// Обработчик для операций с конкретной задачей (по ID)
	http.HandleFunc("/api/tasks/", app.ProtectedApiMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем task ID из URL
		path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи не указан"})
			return
		}

		// Обрабатываем разные методы
		switch r.Method {
		case http.MethodPut:
			// Проверяем, это toggle или обычное обновление
			if len(parts) > 1 && parts[1] == "toggle" {
				app.ToggleTaskStatusHandler(w, r)
			} else {
				app.SaveTaskDataHandler(w, r)
				// Здесь можно добавить обработчик для обычного обновления задачи
				//w.WriteHeader(http.StatusNotImplemented)
				//json.NewEncoder(w).Encode(map[string]string{"error": "Обновление задачи пока не реализовано"})
			}
		case http.MethodDelete:
			app.DeleteTaskHandler(w, r)
		case http.MethodGet:
			app.GetTaskHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		}
	}))

	// Обработчик смены пароля
	http.HandleFunc("/api/user/password", app.ProtectedApiMiddleware(app.SaveUserPasswordHandler))

	// Обработка смены логина
	http.HandleFunc("/api/user/email", app.ProtectedApiMiddleware(app.SaveUserEmailHandler))

	// Страницы
	http.HandleFunc("/", app.RegisterFormHandler)
	http.HandleFunc("/dashboard", app.DashboardHandler)

	port := cfg.GetServerPortString()
	server := &http.Server{
		Addr:    port,
		Handler: nil,
	}

	// Запуск сервера
	go func() {
		log.Printf("Сервер запущен на порту %s\n", port)
		log.Printf("Откройте в браузере: http://localhost%s\n", port)
		log.Printf("Dashboard: http://localhost%s/dashboard\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Инициализируем Kafka Producer
	kafkaProducer, err := services.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaNotificationTopic)
	if err != nil {
		log.Fatalf("Ошибка инициализации Kafka Producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Создаем и запускаем сервис проверки задач
	interval, err := time.ParseDuration(cfg.NotificationCheckInterval)
	if err != nil {
		interval = time.Minute // значение по умолчанию
	}

	taskChecker := services.NewTaskChecker(db, kafkaProducer, interval)
	go taskChecker.Start()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("\nЗавершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении сервера: %v", err)
	}
	log.Println("Сервер корректно завершен")
}
