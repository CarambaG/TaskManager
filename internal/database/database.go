package database

import (
	"TaskManager/internal/config"
	"database/sql"
	"errors"
	"log"
	"time"
)

func ConnectDatabase(cfg *config.Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Пытаемся подключиться несколько раз (для Docker Compose)
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		db, err = sql.Open("postgres", cfg.GetConnectionString())
		if err != nil {
			log.Printf("Попытка %d: ошибка подключения к БД: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Printf("Попытка %d: БД не отвечает: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Успешное подключение к БД")
		return db, nil
	}

	log.Fatal("Не удалось подключиться к БД после нескольких попыток")
	return nil, errors.New("не удалось подключиться к БД")
}

func PingDatabase(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		log.Fatal("Не удалось подключиться к БД.")
		return errors.New("Не удалось подключиться к БД.")
	}
	return nil
}
