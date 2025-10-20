package models

import "time"

type User struct {
	ID       string    `json:"id"`
	Login    string    `json:"login"`
	PassHash string    `json:"-"`
	CreateAt time.Time `json:"create_at"`
}
