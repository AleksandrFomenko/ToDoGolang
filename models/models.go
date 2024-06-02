package models

import (
	"time"
)

type Task struct {
	ID      int       `json:"id"`
	Date    time.Time `json:"date"`
	Title   string    `json:"title"`
	Comment string    `json:"comment"`
	Repeat  string    `json:"repeat"`
}
