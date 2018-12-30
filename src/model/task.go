package model

import (
	"time"
)

type Task struct {
	ID        uint      `json:"id"`         // id
	CreatedAt time.Time `json:"created_at"` // created_at
	UpdatedAt time.Time `json:"updated_at"` // updated_at
	Title     string    `json:"title"`      // title

}
