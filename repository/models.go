package repository

import "time"

type User struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Auth struct {
	SessionID string `json:"sessionId"`

	UserID    int64     `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
