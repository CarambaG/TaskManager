package controllers

import (
	"TaskManager/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateUser(db *sql.DB, login, pass string) (*models.User, error) {
	// Валидация входных данных
	if login == "" || pass == "" {
		return nil, errors.New("логин и пароль не могут быть пустыми")
	}

	//Проверка наличия аккаунта с тем же логином
	var count int
	err := db.QueryRow("SELECT count(login) FROM users WHERE login = $1", login).Scan(&count)
	if err != nil {
		log.Println("Ошибка в проверке наличия")
		return nil, err
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("Пользователь с логином %s уже существует", login))
	}

	//Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), 12) // cost = 12
	if err != nil {
		log.Println("Ошибка в хешировании пароля")
		return nil, err
	}

	//Создаем пользователя
	user := models.User{
		Login:    login,
		PassHash: string(hash),
		CreateAt: time.Now(),
	}

	//Вставляем аккаунт в БД
	err = db.QueryRow("INSERT INTO users (login, pass, create_at) VALUES ($1, $2, $3) RETURNING id",
		user.Login, user.PassHash, user.CreateAt).Scan(&user.ID)
	if err != nil {
		log.Println("Ошибка во вставке аккаунта")
		return nil, err
	}

	return &user, nil
}

func Authenticate(db *sql.DB, login, password string) (*models.User, error) {
	// Валидация
	if login == "" || password == "" {
		return nil, errors.New("логин и пароль не могут быть пустыми")
	}

	var user models.User

	err := db.QueryRow(
		"SELECT id, login, pass FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.PassHash)
	if err == sql.ErrNoRows {
		return nil, errors.New("Пользователь не найден")
	}
	if err != nil {
		return nil, err
	}

	// Проверяем пароль
	if !CheckPasswordHash(password, user.PassHash) {
		return nil, fmt.Errorf("Неверный пароль")
	}

	user.PassHash = "" // Не возвращаем хэш пароля
	return &user, nil
}
