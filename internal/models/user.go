package models

import "time"

type User struct {
	ID       string    `json:"id"`
	Login    string    `json:"login"`
	PassHash string    `json:"-"`
	Email    string    `json:"email"`
	CreateAt time.Time `json:"create_at"`
}
