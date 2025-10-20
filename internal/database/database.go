package database

import (
	"TaskManager/internal/config"
	"database/sql"
	"errors"
	"log"
)

func ConnectDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetConnectionString())
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
		return nil, errors.New("Ошибка подключения к БД")
	}
	return db, nil
}

func PingDatabase(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		log.Fatal("Не удалось подключиться к БД.")
		return errors.New("Не удалось подключиться к БД.")
	}
	return nil
}
