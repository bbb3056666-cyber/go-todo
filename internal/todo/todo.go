package todo

import (
	"errors"
	"time"
)

var ErrInvalidTitle = errors.New("title 必须为 1 到 100 个 Unicode 码点")

// Todo is a single in-memory task.
type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
